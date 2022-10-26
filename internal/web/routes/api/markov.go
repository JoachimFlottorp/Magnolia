package api

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/JoachimFlottorp/magnolia/internal/ctx"
	"github.com/JoachimFlottorp/magnolia/internal/rabbitmq"
	"github.com/JoachimFlottorp/magnolia/internal/web/response"
	"github.com/JoachimFlottorp/magnolia/internal/web/router"
	pb "github.com/JoachimFlottorp/magnolia/protobuf"
	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/gorilla/mux"
)

var (
	ErrNoData 			= "no data"
	ErrTooLong 			= "took to long to generate markov chain"
	ErrUnableToGenerate = "unable to generate markov chain"

	ValidChannel = regexp.MustCompile(`^[a-zA-Z0-9_]{4,25}$`)
)

func ErrNotEnoughData(len int) string {
	return fmt.Sprintf("not enough data (%d/100)", len)
}

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

	markovReqs map[string]chan *pb.MarkovResponse

	isAlive bool
}

func NewMarkovRoute(gCtx ctx.Context) router.Route {
	_, err := gCtx.Inst().RMQ.CreateQueue(gCtx, rabbitmq.QueueSettings{
		Name: rabbitmq.QueueJoinRequest,
	})

	if err != nil {
		zap.S().Fatalw("Failed to create rabbitmq queue", "name", rabbitmq.QueueJoinRequest, "error", err)
	}

	a := &MarkovRoute{
		Ctx: gCtx,
		markovReqs: make(map[string]chan *pb.MarkovResponse),
	}

	go a.handleMarkovRequests()

	a.pingMarkovGenerator()
	go func() {
		for {
			select {
			case <-a.Ctx.Done(): {
				return
			}
			case <-time.After(1 * time.Minute): {
				a.pingMarkovGenerator()
			}
			}
		}
	}()	

	return a
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
	var seed 	string
	
	channel = r.URL.Query().Get("channel")
	key := fmt.Sprintf("twitch:%s:chat-data", channel)
	u := uuid.New()
	
	if channel == "" {
		return response.
			Error().
			BadRequest("Missing channel parameter").
			Build()
	} else if !ValidChannel.MatchString(channel) {
		return response.
			Error().
			BadRequest("Invalid channel name").
			Build()
	}

	if s := r.URL.Query().Get("seed"); s != "" {
		seed = s
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
	} else if l := len(storedData); l < 100 {
		return response.
			Error().
			NotFound(ErrNotEnoughData(l)).
			Build()
	} else if !a.isAlive {
		return response.
			Error().
			InternalServerError("Markov generator is not alive").
			Build()
	}

	markovChan, err := a.genMarkov(r.Context(), u, storedData, seed)

	if err != nil {
		return response.
			Error().
			InternalServerError().
			SetCustomReqID(u).
			Build()
	}

	result, ok := <-markovChan

	if !ok {
		zap.S().Errorw("Took to long to generate markov chain", "channel", channel)

		return response.
			Error().
			InternalServerError(ErrTooLong).
			SetCustomReqID(u).
			Build()
	}

	if result.Error != nil {

		if strings.HasPrefix(*result.Error, "Failed to build a sentence after") {
			return response.
				Error().
				BadRequest(ErrUnableToGenerate).
				SetCustomReqID(u).
				Build()
		}
		
		return response.
			Error().
			InternalServerError().
			SetCustomReqID(u).
			Build()
	}
	
	return response.OkResponse().
		SetJSON(MarkovResponse{ Markov: result.Result }).
		Build()

}

func (a *MarkovRoute) handleMarkovRequests() {
	msg, err := a.Ctx.Inst().RMQ.Consume(a.Ctx, rabbitmq.ConsumeSettings{
		Queue: rabbitmq.QueueMarkovGenenerator,
	})
	
	if err != nil {
		zap.S().Errorw("Failed to consume from RabbitMQ", "error", err)
		return
	}
	
	for {
		select {
		case m := <-msg: {
			var res pb.MarkovResponse
			err := proto.Unmarshal(m.Body, &res)
			if res.Error != nil {
				zap.S().Errorw("Failed to generate markov chain", "error", res.Error)
			}
			if err != nil {
				zap.S().Errorw("Failed to unmarshal protobuf message", "error", err)
				continue
			}

			zap.S().Debugf("Generated markov chain: %s", res.Result)

			if ch, ok := a.markovReqs[m.CorrelationId]; ok {
				ch <- &res
				close(ch)
				delete(a.markovReqs, m.CorrelationId)
			} else {
				a.markovReqs[m.CorrelationId] = make(chan *pb.MarkovResponse)
				a.markovReqs[m.CorrelationId] <- &res
			}
		}
		case <-a.Ctx.Done(): return
		}
	}
}

func (a *MarkovRoute) genMarkov(ctx context.Context, corrId uuid.UUID, data []string, seed string) (chan *pb.MarkovResponse, error) {
	p := pb.MarkovRequest {
		Messages: data,
	}

	if seed != "" {
		p.Seed = &seed
	}

	reqByte, err := proto.Marshal(&p)
	if err != nil { return nil, err }
	
	_ = a.Ctx.Inst().RMQ.Publish(ctx, rabbitmq.PublishSettings{
		RoutingKey: rabbitmq.QueueMarkovGenenerator,
		Msg: amqp091.Publishing{
			CorrelationId: corrId.String(),
			Body: reqByte,
		},
	})

	markovChan := make(chan *pb.MarkovResponse)

	go func() {
		if a.markovReqs[corrId.String()] == nil {
			a.markovReqs[corrId.String()] = make(chan *pb.MarkovResponse)
		}
		
		for {
			select {
			case <-ctx.Done(): return
			case res, ok := <-a.markovReqs[corrId.String()]: {
				if !ok {
					return
				}
				markovChan <- res
				close(markovChan)
				delete(a.markovReqs, corrId.String())
				return
			}
			}
		}
	}()

	return markovChan, nil
}

func (a *MarkovRoute) pingMarkovGenerator() {
	zap.S().Infow("Pinging markov generator")

	url := fmt.Sprintf("%s/health", a.Ctx.Config().Markov.HealthAddress)
	req, err := http.NewRequestWithContext(a.Ctx, http.MethodGet, url, nil)
	if err != nil {
		zap.S().Errorw("Failed to create health check request", "error", err)

		a.isAlive = false
		
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		zap.S().Errorw("Failed to execute health check request", "error", err)

		a.isAlive = false
		
		return
	}

	if resp.StatusCode != http.StatusOK {
		zap.S().Errorw("Markov generator is not healthy", "status", resp.StatusCode)

		a.isAlive = false
		
		return
	}

	a.isAlive = true
	zap.S().Infow("Markov generator is healthy")
}
