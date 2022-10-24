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

	"github.com/JoachimFlottorp/magnolia/internal/config"
	"github.com/JoachimFlottorp/magnolia/internal/ctx"
	"github.com/JoachimFlottorp/magnolia/internal/mongo"
	"github.com/JoachimFlottorp/magnolia/internal/rabbitmq"
	"github.com/JoachimFlottorp/magnolia/internal/redis"
	"github.com/JoachimFlottorp/magnolia/internal/web"

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

	zap.S().Info("Ready!")
	
	<-done

	os.Exit(0)
}
