package api

import (
	"net/http"
	"time"

	"github.com/JoachimFlottorp/magnolia/internal/ctx"
	"github.com/JoachimFlottorp/magnolia/internal/web/router"
	"github.com/JoachimFlottorp/magnolia/internal/web/routes/api/markov"
	"github.com/gofiber/fiber/v2"
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
		Children: []router.RouteInitializerFunc{
			markov.NewGetRoute,
		},
	}
}

func (a *Route) Handler() router.RouterHandler {
	return func(c *fiber.Ctx) (int, interface{}, error) {
		return http.StatusOK, HealthResponse{
			Name:   "api",
			Uptime: int64(t.UnixMilli()),
		}, nil
	}
}
