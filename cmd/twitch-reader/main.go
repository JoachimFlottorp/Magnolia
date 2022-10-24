// TODO Prometheus

package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/JoachimFlottorp/magnolia/cmd/twitch-reader/irc"
	"github.com/JoachimFlottorp/magnolia/internal/config"
	"github.com/JoachimFlottorp/magnolia/internal/ctx"
	"github.com/JoachimFlottorp/magnolia/internal/mongo"
	"github.com/JoachimFlottorp/magnolia/internal/rabbitmq"
	"github.com/JoachimFlottorp/magnolia/internal/redis"
	pb "github.com/JoachimFlottorp/magnolia/protobuf"
	"google.golang.org/protobuf/proto"

	"go.uber.org/zap"
)

var (
	cfg 	= flag.String("config", "config.json", "Path to the config file")
	debug 	= flag.Bool("debug", false, "Enable debug logging")
)

func init() {
	flag.Parse()
	
	if *debug {
		b, _ := zap.NewDevelopmentConfig().Build()
		zap.ReplaceGlobals(b)
	} else {
		b, _ := zap.NewProductionConfig().Build()
		zap.ReplaceGlobals(b)
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
	}();

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
			Address: conf.Redis.Address,
			Username: conf.Redis.Username,
			Password: conf.Redis.Password,
			DB: conf.Redis.Database,
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

		_ = gCtx.Inst().Mongo.RawDatabase().CreateCollection(gCtx, string(mongo.CollectionAPILog))
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
			msg, err := gCtx.Inst().RMQ.Consume(gCtx, rabbitmq.ConsumeSettings {
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
					}
	
					onJoinRequest(gCtx, ircMan, req)
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
				case <-gCtx.Done(): return
				case msg := <-ircMan.MessageQueue: {
					zap.S().Debugw("Received message from irc manager", "message", msg)
				}
			}
		}
	}()

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

	zap.S().Infow("Joining channel", "channel", channel.TwitchName)

	irc.JoinChannel(channel)
}