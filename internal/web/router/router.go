package router

import (
	"github.com/JoachimFlottorp/magnolia/internal/ctx"
	"github.com/gofiber/fiber/v2"
)

type RouteInitializerFunc func(ctx.Context) Route

// RouteConfig: Specifies the configuration of a route.
type RouteConfig struct {
	URI        string
	Method     []string
	Children   []RouteInitializerFunc
	Middleware []Middleware
}

type Route interface {
	Configure() RouteConfig
	Handler() fiber.Handler
}

type Middleware interface {
	Handler() fiber.Handler
}
