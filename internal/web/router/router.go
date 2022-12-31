package router

import (
	"fmt"

	"github.com/JoachimFlottorp/magnolia/internal/ctx"
	"github.com/gofiber/fiber/v2"
)

type RouteInitializerFunc func(ctx.Context) Route
type RouterHandler func(c *fiber.Ctx) (int, interface{}, error)

var (
	ErrInternalServerError = fmt.Errorf("internal Server Error")
)

// RouteConfig: Specifies the configuration of a route.
type RouteConfig struct {
	URI        string
	Method     []string
	Children   []RouteInitializerFunc
	Middleware []Middleware
}

type Route interface {
	Configure() RouteConfig
	Handler() RouterHandler
}

type Middleware interface {
	Handler() fiber.Handler
}
