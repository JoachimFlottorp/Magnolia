package markov

import (
	"net/http"

	"github.com/JoachimFlottorp/magnolia/internal/ctx"
	"github.com/JoachimFlottorp/magnolia/internal/mongo"
	"github.com/JoachimFlottorp/magnolia/internal/web/locals"
	"github.com/JoachimFlottorp/magnolia/internal/web/router"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
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
		URI:    "/list",
		Method: []string{http.MethodGet},
	}
}

// swagger:route GET /api/markov/list markov
//
// Get a list of channels which are currently being tracked.
//
//	Responses:
//		200: MarkovListResponse
func (a *ListRoute) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		cur, err := a.Ctx.Inst().Mongo.Collection(mongo.CollectionTwitch).Find(c.Context(), bson.D{})
		if err != nil {
			zap.S().Errorw("failed to find channels", "error", err)

			c.Locals(locals.LocalStatus, http.StatusInternalServerError)
			return c.Next()
		}

		var resp MarkovListResponse
		for cur.Next(c.Context()) {
			var channel mongo.TwitchChannel
			if err := cur.Decode(&channel); err != nil {
				zap.S().Errorw("failed to decode channel", "error", err)

				c.Locals(locals.LocalStatus, http.StatusInternalServerError)
				return c.Next()
			}

			resp.Channels = append(resp.Channels, basicChannel{
				Username: channel.TwitchName,
				UserID:   channel.TwitchID,
			})
		}

		c.Locals(locals.LocalStatus, http.StatusOK)
		c.Locals(locals.LocalResponse, resp)

		return c.Next()
	}
}
