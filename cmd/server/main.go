// TODO Prometheus

package main

import (
	"context"
	"flag"
	"net/http"
	"sync"

	"github.com/JoachimFlottorp/GoCommon/cron"
	"github.com/JoachimFlottorp/magnolia/external/emotes"
	"github.com/JoachimFlottorp/magnolia/external/emotes/models"
	recentmessages "github.com/JoachimFlottorp/magnolia/external/recent-messages"
	"github.com/JoachimFlottorp/magnolia/internal/config"
	"github.com/JoachimFlottorp/magnolia/internal/ctx"
	"github.com/JoachimFlottorp/magnolia/internal/mongo"
	"github.com/JoachimFlottorp/magnolia/internal/web"
	"github.com/JoachimFlottorp/magnolia/pkg/sigwrapper"

	"go.mongodb.org/mongo-driver/bson"

	"go.uber.org/zap"
)

func init() {
	flag.Parse()
}

func main() {
	conf, err := config.CreateConfig()
	if err != nil {
		panic(err)
	}

	gCtx, cancel, err := ctx.CreateAndPopulateGlobalContext(conf)
	if err != nil {
		zap.S().Fatalw("Failed to create global context", "error", err)
	}

	done := sigwrapper.NewWrapper(gCtx, cancel, zap.S())

	done.Run(func(ctx context.Context) {
		wg := sync.WaitGroup{}

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

			cronMan := cron.NewManager(ctx, false)

			cronMan.Add(cron.CronOptions{
				Name:   "updateRecentMessageBroker",
				Spec:   "*/5 * * * *",
				RunNow: false,
				Cmd:    func() { updateRecentMessageBroker(ctx, gCtx.Inst().Mongo) },
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

					for channels.Next(ctx) {
						var channel mongo.TwitchChannel

						err := channels.Decode(&channel)

						if err != nil {
							zap.S().Errorw("Failed to decode channel", "error", err)
							continue
						}

						e, err := emotes.GetEmotes(ctx, models.ChannelIdentifier{
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

			<-ctx.Done()
		}()
	})
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

	recentmessages.Request(recentmessages.EndpointSnakes, c)

	zap.S().Infof("Requested recent messages for %d channels", len(c))
}
