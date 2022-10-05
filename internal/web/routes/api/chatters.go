package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/JoachimFlottorp/yeahapi/internal/ctx"
	"github.com/JoachimFlottorp/yeahapi/types/twitch"

	"github.com/JoachimFlottorp/yeahapi/internal/web/router"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// swagger:parameters chattersGet
type chattersGetParams struct {
	// The login of the channel
	// in: path
	// description: The login of the channel
	// required: true
	// type: string
	Login string `json:"login"`
}

type cRoute struct {
	Ctx ctx.Context
}

func newChatters(gCtx ctx.Context) router.Route {
	return &cRoute{gCtx}
}

func (a *cRoute) Configure() router.RouteConfig {
	return router.RouteConfig{
		URI: "/{login}/chatters",
		Method: []string{http.MethodGet},
		Children: []router.Route{},
		Middleware: []mux.MiddlewareFunc{},
	}
}

//	swagger:route GET /api/{login}/chatters chatters chattersGet
//
// 	Get the chatters of a channel
//
// 			Responses:
// 				200: Chatters
// 				400: apiFail
// 				500: apiFail
//
func (a *cRoute) Handler(w http.ResponseWriter, r *http.Request) {
	login := mux.Vars(r)["login"]
	if login == "" {
		a.Ctx.ApiErr(w, r, http.StatusBadRequest, errors.New("missing login"))
	}

	cached, err := a.Ctx.Inst().Redis.Get(r.Context(), fmt.Sprintf("channel:%s:chatters", login))
	if err != nil && err != redis.Nil {
		zap.S().Errorw("RedisError", "error", err)
		
		a.Ctx.ApiErr(w, r, http.StatusInternalServerError, errors.New("internal server error"))
		return
	} else if cached != "" {
		c := &twitch.Chatters{}
		
		err := json.Unmarshal([]byte(cached), c)
		if err != nil {
			zap.S().Errorw("JsonError", "error", err)
			
			a.Ctx.ApiErr(w, r, http.StatusInternalServerError, errors.New("internal server error"))
			return
		}
		a.Ctx.ApiOK(w, r, http.StatusOK, c)
		return
	}

	c, err := a.Ctx.Inst().Grpc.ChatterGet(r.Context(), login)
	if err != nil {
		zap.S().Errorw("GrpcError", "error", err)
		
		a.Ctx.ApiErr(w, r, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}

	go func () {
		b, err := json.Marshal(c)
		if err != nil {
			zap.S().Errorw("JsonError", "error", err)
			return
		}

		key 	:= fmt.Sprintf("channel:%s:chatters", login)
		expiry 	:= time.Duration(5) * time.Minute
		
		err = a.Ctx.Inst().Redis.Set(a.Ctx, key, string(b))
		if err != nil {
			zap.S().Errorw("RedisError", "error", err)
			return
		}

		a.Ctx.Inst().Redis.Expire(a.Ctx, key, expiry)
	}()

	a.Ctx.ApiOK(w, r, http.StatusOK, twitch.Chatters{
		Total: int(c.GetTotal()),
		Chatters: c.Chatters,
	})
}
