// TODO Prometheus

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"sync"
	"syscall"
	"time"

	"github.com/JoachimFlottorp/magnolia/internal/config"
	"github.com/JoachimFlottorp/magnolia/internal/ctx"
	"github.com/JoachimFlottorp/magnolia/internal/mongo"
	"github.com/JoachimFlottorp/magnolia/internal/rabbitmq"
	"github.com/JoachimFlottorp/magnolia/internal/redis"
	"github.com/JoachimFlottorp/magnolia/pkg/irc"
	pb "github.com/JoachimFlottorp/magnolia/protobuf"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/protobuf/proto"

	"go.uber.org/zap"
)

var (
	cfg    = flag.String("config", "config.json", "Path to the config file")
	debug  = flag.Bool("debug", false, "Enable debug logging")
	maxMsg = flag.Int64("max-msg", 1000, "Maximum number of messages to store in redis")

	botIgnoreList = regexp.MustCompile(`bo?t{1,2}(?:(?:ard)?o|\d|_)*$|^(?:fembajs|veryhag|scriptorex|apulxd|qdc26534|linestats|pepegaboat|sierrapine|charlestonbieber|icecreamdatabase|chatvote|localaniki|rewardmore|gorenmu|0weebs|befriendlier|electricbodybuilder|o?bot(?:bear1{3}0|2465|menti|e|nextdoor)|stream(?:elements|labs))$`)
)

func init() {
	flag.Parse()

	if err := config.ReplaceZapGlobal(*debug); err != nil {
		panic(err)
	}

	if cfg == nil {
		zap.S().Fatal("Config file is not set")
	}
}

func main() {
	cfgFile, err := os.OpenFile(*cfg, os.O_RDONLY, 0)
	if err != nil {
		zap.S().Fatalw("Config file is not set", "error", err)
	}

	defer func() {
		err := cfgFile.Close()
		zap.S().Warnw("Failed to close config file", "error", err)
	}()

	conf := &config.Config{}
	err = json.NewDecoder(cfgFile).Decode(conf)
	if err != nil {
		zap.S().Fatalw("Failed to decode config file", "error", err)
	}

	doneSig := make(chan os.Signal, 1)
	signal.Notify(doneSig, syscall.SIGINT, syscall.SIGTERM)

	gCtx, cancel := ctx.WithCancel(ctx.New(context.Background(), conf))

	ircMan := irc.NewManager(gCtx)

	{
		gCtx.Inst().Redis, err = redis.Create(gCtx, redis.Options{
			Address:  conf.Redis.Address,
			Username: conf.Redis.Username,
			Password: conf.Redis.Password,
			DB:       conf.Redis.Database,
		})

		if err != nil {
			zap.S().Fatalw("Failed to create redis instance", "error", err)
		}
	}

	{
		gCtx.Inst().Mongo, err = mongo.New(gCtx, conf)

		if err != nil {
			zap.S().Fatalw("Failed to create mongo instance", "error", err)
		}
	}

	{
		gCtx.Inst().RMQ, err = rabbitmq.New(gCtx, &rabbitmq.NewInstanceSettings{
			Address: gCtx.Config().RabbitMQ.URI,
		})

		if err != nil {
			zap.S().Fatalw("Failed to create rabbitmq instance", "error", err)
		}

		if _, err = gCtx.Inst().RMQ.CreateQueue(gCtx, rabbitmq.QueueSettings{
			Name: rabbitmq.QueueJoinRequest,
		}); err != nil {
			zap.S().Fatalw("Failed to create rabbitmq queue", "error", err)
		}

		go func() {
			msg, err := gCtx.Inst().RMQ.Consume(gCtx, rabbitmq.ConsumeSettings{
				Queue: rabbitmq.QueueJoinRequest,
			})
			if err != nil {
				zap.S().Fatalw("Failed to consume rabbitmq queue", "error", err)
			}
			for {
				select {
				case <-gCtx.Done():
					return
				case m := <-msg:
					req := &pb.SubChannelReq{}
					err = proto.Unmarshal(m.Body, req)
					if err != nil {
						zap.S().Fatalw("Failed to unmarshal rabbitmq message", "error", err)
						continue
					}

					onJoinRequest(gCtx, ircMan, req)
				}
			}
		}()

		go func() {
			msg, err := gCtx.Inst().RMQ.Consume(gCtx, rabbitmq.ConsumeSettings{
				Queue: rabbitmq.QueuePartRequest,
			})

			if err != nil {
				zap.S().Fatalw("Failed to consume rabbitmq queue", "error", err)
			}
			for {
				select {
				case <-gCtx.Done():
					return
				case m := <-msg:
					req := &pb.SubChannelReq{}
					err = proto.Unmarshal(m.Body, req)
					if err != nil {
						zap.S().Fatalw("Failed to unmarshal rabbitmq message", "error", err)
						continue
					}

					ircMan.LeaveChannel(req.Channel)

					_, err = gCtx.Inst().Mongo.
						Collection(mongo.CollectionTwitch).
						DeleteOne(gCtx, bson.M{
							"twitch_name": req.Channel,
						})

					if err != nil {
						zap.S().Errorw("Failed to delete channel from mongo", "error", err)
					}
				}
			}
		}()
	}

	wg := sync.WaitGroup{}

	done := make(chan any)

	go func() {
		<-doneSig
		cancel()

		go func() {
			select {
			case <-time.After(10 * time.Second):
			case <-doneSig:
			}
			zap.S().Fatal("Forced to shutdown, because the shutdown took too long")
		}()

		zap.S().Info("Shutting down")

		wg.Wait()

		zap.S().Info("Shutdown complete")
		close(done)
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()

		err = ircMan.ConnectAllFromDatabase()
		if err != nil {
			zap.S().Fatalw("Failed to setup irc manager", "error", err)
		}

		for {
			select {
			case <-gCtx.Done():
				return
			case msg := <-ircMan.MessageQueue:
				{
					key := fmt.Sprintf("twitch:%s:chat-data", msg.Channel)
					data := msg.Message
					user := msg.User
					if botIgnoreList.MatchString(user) {
						continue
					}

					if len, err := gCtx.Inst().Redis.LLen(gCtx, key); err != nil {
						zap.S().Errorw("Failed to get length of redis list", "error", err)
						continue
					} else {
						if len >= *maxMsg {
							if err := gCtx.Inst().Redis.LRPop(gCtx, key); err != nil {
								zap.S().Errorw("Failed to pop redis list", "error", err)
								continue
							}
						}
					}

					if err := gCtx.Inst().Redis.LPush(gCtx, key, data); err != nil {
						zap.S().Errorw("Failed to push message to redis", "error", err)
						continue
					}

					pbMsg := pb.IRCPrivmsg{
						Message: data,
						Channel: msg.Channel,
						User: &pb.IRCUser{
							Username: user,
							UserId:   msg.Tags["user-id"],
						},
					}

					msgByt, err := proto.Marshal(&pbMsg)

					if err != nil {
						zap.S().Errorw("Failed to marshal protobuf message", "error", err)
						continue
					}

					if err := gCtx.Inst().Redis.Publish(gCtx, "twitch:messages", msgByt); err != nil {
						zap.S().Errorw("Failed to publish message to redis", "error", err)
						continue
					}
				}
			}
		}
	}()

	// wg.Add(1)

	// go func() {
	// 	defer wg.Done()

	// 	queryUpdateBotList(gCtx)
	// }()

	zap.S().Info("Ready!")

	<-done

	os.Exit(0)
}

func onJoinRequest(gCtx ctx.Context, irc *irc.IrcManager, req *pb.SubChannelReq) {
	if req.Channel == "" {
		return
	}

	channel := mongo.TwitchChannel{
		TwitchName: req.Channel,
	}

	if err := channel.GetByName(gCtx, gCtx.Inst().Mongo); err == mongo.ErrNoDocuments {
		err = channel.ResolveByIVR(gCtx)
		if err != nil {
			zap.S().Errorw("Failed to resolve channel by IVR", "error", err)
		}

		channel.Save(gCtx, gCtx.Inst().Mongo)
	}

	irc.JoinChannel(channel)
}

// func queryUpdateBotList(ctx context.Context) {
// 	for {
// 		select {
// 		case <-ctx.Done(): return
// 		case <-time.After(1 * time.Minute): {
// 			botList, err := supibot.GetBotList(ctx, external.Client())

// 			if err != nil {
// 				zap.S().Errorw("Failed to update bot list", "error", err)
// 				continue
// 			}

// 			var newList []string
// 			for _, bot := range botList.Data.Bots {
// 				newList = append(newList, bot.Name)
// 			}

// 			newList = append(newList, ignoreBots...)
// 			updateBotList(newList)

// 			zap.S().Debugw("Updated bot list", "count", len(newList), "list", botIgnoreList.String())
// 		}
// 		}
// 	}
// }

// func updateBotList(bots []string) {
// 	botIgnoreList = regexp.MustCompile(fmt.Sprintf("(%s)", strings.Join(bots, "|")))
// }
