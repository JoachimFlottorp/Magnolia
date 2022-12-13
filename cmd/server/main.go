// TODO Prometheus

package main

import (
	"context"
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/JoachimFlottorp/GoCommon/cron"
	"github.com/JoachimFlottorp/magnolia/external/emotes"
	"github.com/JoachimFlottorp/magnolia/external/emotes/models"
	recentmessages "github.com/JoachimFlottorp/magnolia/external/recent-messages"
	"github.com/JoachimFlottorp/magnolia/internal/config"
	"github.com/JoachimFlottorp/magnolia/internal/ctx"
	"github.com/JoachimFlottorp/magnolia/internal/mongo"
	"github.com/JoachimFlottorp/magnolia/internal/rabbitmq"
	"github.com/JoachimFlottorp/magnolia/internal/redis"
	"github.com/JoachimFlottorp/magnolia/internal/web"

	"go.mongodb.org/mongo-driver/bson"

	"go.uber.org/zap"
)

var (
	cfg   = flag.String("config", "config.json", "Path to the config file")
	debug = flag.Bool("debug", false, "Enable debug logging")
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

		if err := web.New(gCtx); err != nil && err != http.ErrServerClosed {
			zap.S().Fatalw("Failed to start web server", "error", err)
		}
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()

		cronMan := cron.NewManager(gCtx, false)

		cronMan.Add(cron.CronOptions{
			Name:   "updateRecentMessageBroker",
			Spec:   "*/5 * * * *",
			RunNow: false,
			Cmd:    func() { updateRecentMessageBroker(gCtx, gCtx.Inst().Mongo) },
		})

		cronMan.Add(cron.CronOptions{
			Name:   "UpdateEmotes",
			Spec:   "*/10 * * * *",
			RunNow: false,
			Cmd: func() {
				channels, err := gCtx.Inst().Mongo.Collection(mongo.CollectionTwitch).Find(gCtx, bson.M{})

				if err != nil {
					zap.S().Errorw("Failed to get channels", "error", err)
					return
				}

				for channels.Next(gCtx) {
					var channel mongo.TwitchChannel

					err := channels.Decode(&channel)

					if err != nil {
						zap.S().Errorw("Failed to decode channel", "error", err)
						continue
					}

					e, err := emotes.GetEmotes(gCtx, models.ChannelIdentifier{
						ID:   channel.TwitchID,
						Name: channel.TwitchName,
					})

					if err != nil {
						zap.S().Errorw("Failed to get emotes", "error", err)
						continue
					}

					e.Save(gCtx, gCtx.Inst().Redis, channel.TwitchName)

					zap.S().Infow("Updated emotes", "channel", channel.TwitchName)
				}
			},
		})

		cronMan.Start()

		<-gCtx.Done()
	}()

	zap.S().Info("Ready!")

	<-done

	os.Exit(0)
}

func updateRecentMessageBroker(ctx context.Context, m mongo.Instance) {
	var channels []mongo.TwitchChannel
	cursor, err := m.Collection(mongo.CollectionTwitch).Find(ctx, bson.M{})
	if err != nil {
		zap.S().Errorw("Failed to get twitch channels", "error", err)
		return
	}

	if err := cursor.All(ctx, &channels); err != nil {
		zap.S().Errorw("Failed to decode twitch channels", "error", err)
		return
	}

	var c []string

	for _, channel := range channels {
		c = append(c, channel.TwitchName)
	}

	if err := recentmessages.Request(recentmessages.EndpointSnakes, c); err != nil {
		zap.S().Errorw("Failed to request recent messages", "error", err)
		return
	}

	zap.S().Infof("Requested recent messages for %d channels", len(c))
}
