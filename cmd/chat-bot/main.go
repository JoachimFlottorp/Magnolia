package main

import (
	"context"
	"sync"

	"github.com/JoachimFlottorp/magnolia/cmd/chat-bot/bot"
	"github.com/JoachimFlottorp/magnolia/internal/config"
	"github.com/JoachimFlottorp/magnolia/internal/ctx"
	"github.com/JoachimFlottorp/magnolia/pkg/sigwrapper"

	"go.uber.org/zap"
)

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
		botConn, err := bot.NewBot(gCtx, bot.Credentials{
			Username: gCtx.Config().Twitch.Bot.Username,
			Password: gCtx.Config().Twitch.Bot.Password,
		})

		if err != nil {
			zap.S().Fatalw("Failed to create bot instance", "error", err)
		}

		wg := sync.WaitGroup{}

		wg.Add(1)

		go func() {
			defer wg.Done()

			err = botConn.Run()

			if err != nil {
				zap.S().Fatalw("Failed to run bot", "error", err)
			}

			<-gCtx.Done()
		}()
	})
}
