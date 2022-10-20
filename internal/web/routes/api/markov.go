package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/JoachimFlottorp/magnolia/internal/ctx"
	"github.com/JoachimFlottorp/magnolia/internal/web/response"
	"github.com/JoachimFlottorp/magnolia/internal/web/router"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/gorilla/mux"
	"github.com/mb-14/gomarkov"
)

// swagger:model MarkovResponse
type MarkovResponse struct {
	// The generated markov chain
	// in: body
	Markov string `json:"markov"`
}

// swagger:parameters markovGet
type MarkovGetParams struct {
	// in: query
	// description: The channel to generate a markov chain from
	// required: true
	// type: string
	Shannel string `json:"channel"`
	// in: query
	// description: Generate a markov chain based on a custom seed
	// required: false
	// type: string
	Seed 	string `json:"seed"`
}

type MarkovRoute struct {
	Ctx ctx.Context
}

func NewMarkovRoute(gCtx ctx.Context) router.Route {
	return &MarkovRoute{gCtx}
}

func (a *MarkovRoute) Configure() router.RouteConfig {
	return router.RouteConfig{
		URI: "/markov",
		Method: []string{http.MethodGet},
		Children: []router.Route{},
		Middleware: []mux.MiddlewareFunc{},
	}
}

// swagger:route GET /api/markov markov markovGet
// 
// Generate a markov chain based on a channel
// 
//			Responses:
//				200: MarkovResponse
func (a *MarkovRoute) Handler(w http.ResponseWriter, r *http.Request) response.RouterResponse {
	var channel string
	seed := []string{gomarkov.StartToken}
	
	channel = r.URL.Query().Get("channel")
	
	if channel == "" {
		return response.
			Error().
			BadRequest("Missing channel parameter").
			Build()
	}

	if s := r.URL.Query().Get("seed"); s != "" {
		seed = append(seed, strings.Split(s, " ")...)
	}

	// TODO: Find a better way of storing data
	storedData, err := a.Ctx.Inst().Redis.Get(r.Context(), fmt.Sprintf("channel:%s:chat-data", channel))
	if err != nil {
		if err == redis.Nil {
			return response.
				Error().
				NotFound("No data found for channel").
				Build()
		}
		
		zap.S().Errorf("Failed to get channel data from redis: %s", err)
		
		return response.
			Error().
			InternalServerError().
			Build()
	}

	var arrayData []string
	if err := json.Unmarshal([]byte(storedData), &arrayData); err != nil {
		zap.S().Errorf("Failed to unmarshal channel data: %s", err)

		return response.
			Error().
			InternalServerError().
			Build()
	}

	chain := gomarkov.NewChain(1)

	for i := range arrayData {
		chain.Add(strings.Split(arrayData[i], " "))
	}
	
	result, err := chain.Generate(seed)
		
	if err != nil {
		zap.S().Errorf("Error generating markov chain: %s", err)
		
		return response.
			Error().
			InternalServerError().
			Build()
	}

	return response.OkResponse().
		SetJSON(MarkovResponse{ Markov: result }).
		Build()
}
