package ctx

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/JoachimFlottorp/yeahapi/internal/config"
	"github.com/JoachimFlottorp/yeahapi/internal/instance"
	"github.com/JoachimFlottorp/yeahapi/internal/mongo"
	"github.com/JoachimFlottorp/yeahapi/internal/web/router"

	"github.com/google/uuid"
)

type Context interface {
	context.Context
	Config() *config.Config
	Inst() *instance.InstanceList
	ApiOK(http.ResponseWriter, *http.Request, int, interface{})
	ApiErr(http.ResponseWriter, *http.Request, int, error)
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

func (g *gCtx) ApiOK(w http.ResponseWriter, r *http.Request, statusCode int, data interface{}) {
	u := uuid.New()
	t := time.Now()

	res := router.ApiOkay{
		Success: true,
		RequestID: u,
		Timestamp: t,
		Data: data,
	}

	go func() {
		s := mongo.ApiLog {
			ID: u.String(),
			Timestamp: t,
			Method: r.Method,
			Path: r.URL.Path,
			Status: statusCode,
			IP: fmt.Sprintf("%s (%s)", r.Header.Get("X-Forwarded-For"), r.RemoteAddr),
			UserAgent: r.UserAgent(),
		}

		g.Inst().Mongo.Collection(mongo.CollectionAPILog).InsertOne(g, s)
	}()
	
	router.Send(w, statusCode, res)
}

func (g *gCtx) ApiErr(w http.ResponseWriter, r *http.Request, statusCode int, err error) {
	u := uuid.New()
	t := time.Now()

	res := router.ApiFail{
		Success: false,
		RequestID: u,
		Timestamp: t,
		Error: err.Error(),
	}
	
	go func() {
		s := mongo.ApiLog {
			ID: u.String(),
			Timestamp: t,
			Method: r.Method,
			Path: r.URL.Path,
			Status: statusCode,
			IP: fmt.Sprintf("%s (%s)", r.Header.Get("X-Forwarded-For"), r.RemoteAddr),
			UserAgent: r.UserAgent(),
			Error: err.Error(),
		}

		g.Inst().Mongo.Collection(mongo.CollectionAPILog).InsertOne(g, s)
	}()
	
	router.Send(w, statusCode, res)
}