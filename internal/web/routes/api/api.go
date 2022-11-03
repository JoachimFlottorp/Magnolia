package api

import (
	"net/http"
	"time"

	"github.com/JoachimFlottorp/magnolia/internal/ctx"
	"github.com/JoachimFlottorp/magnolia/internal/web/response"
	"github.com/JoachimFlottorp/magnolia/internal/web/router"

	"github.com/gorilla/mux"
)

var t = time.Now()

type HealthResponse struct {
	Name   string `json:"name"`
	Uptime int64  `json:"uptime"`
}

type Route struct {
	Ctx ctx.Context
}

func NewApi(gCtx ctx.Context) router.Route {
	return &Route{gCtx}
}

func (a *Route) Configure() router.RouteConfig {
	return router.RouteConfig{
		URI:    "/api",
		Method: []string{http.MethodGet},
		Children: []router.Route{
			NewMarkovRoute(a.Ctx),
		},
		Middleware: []mux.MiddlewareFunc{},
	}
}

func (a *Route) Handler(w http.ResponseWriter, r *http.Request) response.RouterResponse {
	return response.
		OkResponse().
		SetJSON(HealthResponse{
			Name:   "api",
			Uptime: int64(t.UnixMilli()),
		}).
		Build()
}
