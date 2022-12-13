package markov

import (
	"net/http"

	"github.com/JoachimFlottorp/magnolia/internal/ctx"
	"github.com/JoachimFlottorp/magnolia/internal/mongo"
	"github.com/JoachimFlottorp/magnolia/internal/web/response"
	"github.com/JoachimFlottorp/magnolia/internal/web/router"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"

	"github.com/gorilla/mux"
)

// swagger:response MarkovListResponse
type MarkovListResponse struct {
	Channels []basicChannel `json:"channels"`
}

type basicChannel struct {
	Username string `json:"username"`
	UserID   string `json:"user_id"`
}

type ListRoute struct {
	Ctx ctx.Context
}

func NewListRoute(gCtx ctx.Context) router.Route {
	return &ListRoute{gCtx}
}

func (a *ListRoute) Configure() router.RouteConfig {
	return router.RouteConfig{
		URI:        "/list",
		Method:     []string{http.MethodGet},
		Children:   []router.Route{},
		Middleware: []mux.MiddlewareFunc{},
	}
}

// swagger:route GET /api/markov/list markov
//
// Get a list of channels which are currently being tracked.
//
//	Responses:
//		200: MarkovListResponse
func (a *ListRoute) Handler(w http.ResponseWriter, r *http.Request) response.RouterResponse {
	cur, err := a.Ctx.Inst().Mongo.Collection(mongo.CollectionTwitch).Find(r.Context(), bson.D{})
	if err != nil {
		zap.S().Errorw("failed to find channels", "error", err)

		return response.
			Error().
			InternalServerError().
			Build()
	}

	var resp MarkovListResponse
	for cur.Next(r.Context()) {
		var channel mongo.TwitchChannel
		if err := cur.Decode(&channel); err != nil {
			zap.S().Errorw("failed to decode channel", "error", err)

			return response.
				Error().
				InternalServerError().
				Build()
		}

		resp.Channels = append(resp.Channels, basicChannel{
			Username: channel.TwitchName,
			UserID:   channel.TwitchID,
		})
	}

	return response.
		OkResponse().
		SetJSON(resp).
		Build()
}
