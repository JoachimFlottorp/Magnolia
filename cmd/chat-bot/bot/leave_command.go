package bot

import (
	"strings"

	"github.com/JoachimFlottorp/magnolia/cmd/chat-bot/bot/cmdctx"
	"github.com/JoachimFlottorp/magnolia/cmd/chat-bot/bot/execlevel"
	"github.com/JoachimFlottorp/magnolia/internal/ctx"
	"github.com/JoachimFlottorp/magnolia/internal/mongo"
	"github.com/JoachimFlottorp/magnolia/internal/rabbitmq"
	pb "github.com/JoachimFlottorp/magnolia/protobuf"
	"github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
)

type leaveCommand struct {
	Ctx ctx.Context
}

func newLeaveCommand(gCtx ctx.Context) Command {
	return leaveCommand{
		Ctx: gCtx,
	}
}

func (c leaveCommand) Name() string {
	return "leave"
}

func (c leaveCommand) ExecutionLevel() execlevel.ExecutionLevel {
	return execlevel.ExecutionLevelAdmin
}

func (c leaveCommand) Execute(ctx cmdctx.Context, b Bot, args []string) error {
	if len(args) == 0 || args[0] == "" {
		b.Say(ctx.Channel(), "Provide a channel FeelsDankMan")
		return nil
	}

	channel := strings.ToLower(args[0])

	mongoChannel := mongo.TwitchChannel{
		TwitchName: channel,
	}

	err := mongoChannel.GetByName(c.Ctx, c.Ctx.Inst().Mongo)
	if err == nil {
		req := pb.SubChannelReq{
			Channel: channel,
		}

		reqByte, err := proto.Marshal(&req)
		if err != nil {
			return err
		}

		err = c.Ctx.Inst().RMQ.Publish(c.Ctx, rabbitmq.PublishSettings{
			RoutingKey: rabbitmq.QueuePartRequest,
			Msg: amqp091.Publishing{
				Body:        reqByte,
				ContentType: "application/protobuf; twitch.SubChannelReq",
			},
		})

		if err != nil {
			return err
		}

		b.Say(ctx.Channel(), "Leaving channel "+channel+" FeelsDankMan")
		return nil
	} else if err != mongo.ErrNoDocuments {
		b.Say(ctx.Channel(), "Error FeelsDankMan")
		return err
	} else if err == mongo.ErrNoDocuments {
		b.Say(ctx.Channel(), "Channel already parted FeelsDankMan")
	}

	return nil
}
