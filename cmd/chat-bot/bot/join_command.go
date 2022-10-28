package bot

import (
	"github.com/JoachimFlottorp/magnolia/cmd/chat-bot/bot/cmdctx"
	"github.com/JoachimFlottorp/magnolia/cmd/chat-bot/bot/execlevel"
	"github.com/JoachimFlottorp/magnolia/internal/ctx"
	"github.com/JoachimFlottorp/magnolia/internal/mongo"
	"github.com/JoachimFlottorp/magnolia/internal/rabbitmq"
	pb "github.com/JoachimFlottorp/magnolia/protobuf"
	"github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
)

type joinCommand struct {
	Ctx ctx.Context
}

func newJoinCommand(gCtx ctx.Context) Command {
	return joinCommand{
		Ctx: gCtx,
	}
}

func (c joinCommand) Name() string {
	return "join"
}

func (c joinCommand) ExecutionLevel() execlevel.ExecutionLevel {
	return execlevel.ExecutionLevelAdmin
}

func (c joinCommand) Execute(ctx cmdctx.Context, b Bot, args []string) error {
	channel := args[0]
	if channel == "" {
		b.Say(ctx.Channel(), "Provide a channel FeelsDankMan")
	}
	
	mongoChannel := mongo.TwitchChannel {
		TwitchName: channel,
	}

	err := mongoChannel.GetByName(c.Ctx, c.Ctx.Inst().Mongo)
	if err == nil {
		b.Say(ctx.Channel(), "Channel already joined FeelsDankMan")
		return nil
	} else if err != mongo.ErrNoDocuments {
		return err
	} else if err == mongo.ErrNoDocuments {
		req := pb.SubChannelReq {
			Channel: channel,
		}

		reqByte, err := proto.Marshal(&req)
		if err != nil {
			return err
		}

		err = c.Ctx.Inst().RMQ.Publish(c.Ctx, rabbitmq.PublishSettings{
			RoutingKey: rabbitmq.QueueJoinRequest,
			Msg: amqp091.Publishing{
				Body: reqByte,
				ContentType: "application/protobuf; twitch.SubChannelReq",
			},
		})
		
		if err != nil {
			return err
		}

		b.Say(ctx.Channel(), "Joining channel " + channel + " FeelsDankMan")
	}

	return nil
}
