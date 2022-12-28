package ctx

import (
	"context"
	"fmt"
	"time"

	"github.com/JoachimFlottorp/magnolia/internal/config"
	"github.com/JoachimFlottorp/magnolia/internal/instance"
)

type Context interface {
	context.Context
	Config() *config.Config
	Inst() *instance.InstanceList
}

type gCtx struct {
	context.Context
	config *config.Config
	inst   *instance.InstanceList
}

func (g *gCtx) Config() *config.Config {
	return g.config
}

func (g *gCtx) Inst() *instance.InstanceList {
	return g.inst
}

func New(ctx context.Context, config *config.Config) Context {
	return &gCtx{
		Context: ctx,
		config:  config,
		inst:    &instance.InstanceList{},
	}
}

func WithCancel(ctx Context) (Context, context.CancelFunc) {
	cfg := ctx.Config()
	inst := ctx.Inst()

	c, cancel := context.WithCancel(ctx)

	return &gCtx{
		Context: c,
		config:  cfg,
		inst:    inst,
	}, cancel
}

func WithDeadline(ctx Context, deadline time.Time) (Context, context.CancelFunc) {
	cfg := ctx.Config()
	inst := ctx.Inst()

	c, cancel := context.WithDeadline(ctx, deadline)

	return &gCtx{
		Context: c,
		config:  cfg,
		inst:    inst,
	}, cancel
}

func WithValue(ctx Context, key interface{}, value interface{}) Context {
	cfg := ctx.Config()
	inst := ctx.Inst()

	return &gCtx{
		Context: context.WithValue(ctx, key, value),
		config:  cfg,
		inst:    inst,
	}
}

func WithTimeout(ctx Context, timeout time.Duration) (Context, context.CancelFunc) {
	cfg := ctx.Config()
	inst := ctx.Inst()

	c, cancel := context.WithTimeout(ctx, timeout)

	return &gCtx{
		Context: c,
		config:  cfg,
		inst:    inst,
	}, cancel
}

func CreateAndPopulateGlobalContext(conf *config.Config) (Context, context.CancelFunc, error) {
	ctx := context.Background()

	gCtx, cancel := WithCancel(New(ctx, conf))

	{
		var err error
		gCtx.Inst().Redis, err = instance.CreateRedisInstance(gCtx, conf)
		if err != nil {
			return nil, cancel, fmt.Errorf("CreateRedisInstance %w", err)
		}
	}

	{
		var err error
		gCtx.Inst().Mongo, err = instance.CreateMongoInstance(gCtx, conf)
		if err != nil {
			return nil, cancel, fmt.Errorf("CreateMongoInstance %w", err)
		}
	}

	{
		var err error
		gCtx.Inst().RMQ, err = instance.CreateRabbitMQInstance(gCtx, conf)
		if err != nil {
			return nil, cancel, fmt.Errorf("CreateRabbitMQInstance %w", err)
		}
	}

	return gCtx, cancel, nil
}
