package sigwrapper

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

type Sigwrapper struct {
	ctx    context.Context
	canc   context.CancelFunc
	sig    chan os.Signal
	logger *zap.SugaredLogger
}

func NewWrapper(ctx context.Context, canc context.CancelFunc, logger *zap.SugaredLogger) *Sigwrapper {
	sig := make(chan os.Signal, 1)

	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	return &Sigwrapper{
		ctx:    ctx,
		canc:   canc,
		logger: logger,
		sig:    sig,
	}
}

func (s *Sigwrapper) Run(fn func(context.Context)) {
	done := make(chan any)

	s.logger.Info("Starting")

	fn(s.ctx)

	<-s.sig
	s.canc()

	s.logger.Info("Received shutdown signal, waiting for all goroutines to finish")

	go func() {
		<-time.After(10 * time.Second)
		s.logger.Error("Forced to shutdown, because the shutdown took too long")
		os.Exit(1)
	}()

	close(done)

	s.logger.Info("Shutdown complete")

	os.Exit(0)
}
