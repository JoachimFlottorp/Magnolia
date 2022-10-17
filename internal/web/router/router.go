package router

import (
	"net/http"

	"github.com/JoachimFlottorp/yeahapi/internal/web/response"
	"github.com/gorilla/mux"
)

// RouteConfig: Specifies the configuration of a route.
type RouteConfig struct {
	URI        string
	Method     []string
	Children   []Route
	Middleware []mux.MiddlewareFunc
}

type Route interface {
	Configure() RouteConfig
	Handler(http.ResponseWriter, *http.Request) response.RouterResponse
}
