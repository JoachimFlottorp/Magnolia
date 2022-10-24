package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/JoachimFlottorp/magnolia/internal/ctx"
	"github.com/JoachimFlottorp/magnolia/internal/rabbitmq"
	"github.com/JoachimFlottorp/magnolia/internal/web/response"
	"github.com/JoachimFlottorp/magnolia/internal/web/router"
	pb "github.com/JoachimFlottorp/magnolia/protobuf"
	"github.com/gogo/protobuf/proto"
	"github.com/rabbitmq/amqp091-go"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/gorilla/mux"
	"github.com/mb-14/gomarkov"
)

var (
	ErrNoData = "no data"
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
	_, err := gCtx.Inst().RMQ.CreateQueue(gCtx, rabbitmq.QueueSettings{
		Name: rabbitmq.QueueJoinRequest,
	})

	if err != nil {
		zap.S().Fatalw("Failed to create rabbitmq queue", "name", rabbitmq.QueueJoinRequest, "error", err)
	}

	return &MarkovRoute{
		Ctx: gCtx,
	}
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
	key := fmt.Sprintf("twitch:%s:chat-data", channel)
	
	if channel == "" {
		return response.
			Error().
			BadRequest("Missing channel parameter").
			Build()
	}

	if s := r.URL.Query().Get("seed"); s != "" {
		seed = append(seed, strings.Split(s, " ")...)
	}

	storedData, err := a.Ctx.Inst().Redis.GetAllList(r.Context(), key)
	if err != nil {
		if err != redis.Nil {
			zap.S().Errorf("Failed to get channel data from redis: %s", err)
		
			return response.
				Error().
				InternalServerError().
				Build()
		}
		
		req := pb.SubChannelReq{
			Channel: channel,
		}

		reqByte, err := proto.Marshal(&req)
		if err != nil {
			zap.S().Errorw("Failed to marshal protobuf message", "error", err)
			
			return response.
				Error().
				InternalServerError().
				Build()
		}
		
		err = a.Ctx.Inst().RMQ.Publish(a.Ctx, rabbitmq.PublishSettings{
			RoutingKey: rabbitmq.QueueJoinRequest,
			Msg: amqp091.Publishing{
				Body: reqByte,
				ContentType: "application/protobuf; twitch.SubChannelReq",
			},
		})

		if err != nil {
			zap.S().Errorw("Failed to send subcribe to RabbitMQ", "error", err)

			return response.
				Error().
				InternalServerError("Chat logger not available").
				Build()
		}
		
		return response.
			Error().
			NotFound(ErrNoData).
			Build()
	}

	chain := gomarkov.NewChain(1)

	for i := range storedData {
		chain.Add(strings.Split(storedData[i], " "))
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
