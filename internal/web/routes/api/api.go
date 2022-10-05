package api

import (
	"net/http"

	"github.com/JoachimFlottorp/yeahapi/internal/ctx"
	"github.com/JoachimFlottorp/yeahapi/internal/web/router"

	"github.com/gorilla/mux"
)

type Route struct {
	Ctx ctx.Context
}

func NewApi(gCtx ctx.Context) router.Route {
	return &Route{gCtx}
}

func (a *Route) Configure() router.RouteConfig {
	return router.RouteConfig{
		URI: "/api",
		Method: []string{http.MethodGet},
		Children: []router.Route{
			newChatters(a.Ctx),
		},
		Middleware: []mux.MiddlewareFunc{},
	}
}

func (a *Route) Handler(w http.ResponseWriter, r *http.Request) {
	m := "This is the API Root. Open the API Documentation located at / for more information"

	a.Ctx.ApiOK(w, r, http.StatusOK, m)
}
