package markov

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/JoachimFlottorp/GoCommon/assert"
	"github.com/JoachimFlottorp/GoCommon/cron"
	"github.com/JoachimFlottorp/magnolia/internal/ctx"
	"github.com/JoachimFlottorp/magnolia/internal/rabbitmq"
	"github.com/JoachimFlottorp/magnolia/internal/web/locals"
	"github.com/JoachimFlottorp/magnolia/internal/web/router"
	pb "github.com/JoachimFlottorp/magnolia/protobuf"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

var (
	ErrNoData           = fmt.Errorf("no data")
	ErrTooLong          = fmt.Errorf("took to long to generate markov chain")
	ErrUnableToGenerate = fmt.Errorf("unable to generate markov chain")
	ErrMarkovNotAlive   = fmt.Errorf("markov generator is not alive")
)

func ErrNotEnoughData(len int) error {
	return fmt.Errorf("not enough data (%d/100)", len)
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
	Seed string `json:"seed"`
}

type MarkovRoute struct {
	Ctx        ctx.Context
	markovReqs map[string]chan *pb.MarkovResponse
	isAlive    bool
	cronMan    *cron.Manager
}

func NewGetRoute(gCtx ctx.Context) router.Route {
	_, err := gCtx.Inst().RMQ.CreateQueue(gCtx, rabbitmq.QueueSettings{
		Name: rabbitmq.QueueJoinRequest,
	})

	if err != nil {
		zap.S().Fatalw("Failed to create rabbitmq queue", "name", rabbitmq.QueueJoinRequest, "error", err)
	}

	a := &MarkovRoute{
		Ctx:        gCtx,
		markovReqs: make(map[string]chan *pb.MarkovResponse),
	}

	a.cronMan = cron.NewManager(gCtx, false)

	go a.handleMarkovRequests()

	assert.Error(a.cronMan.Add(cron.CronOptions{
		Name:   "ping_markov_generator",
		Spec:   "*/5 * * * *",
		RunNow: true,
		Cmd: func() {
			a.pingMarkovGenerator()
		},
	}), "Failed to add cron job")

	a.cronMan.Start()

	return a
}

func (a *MarkovRoute) Configure() router.RouteConfig {
	return router.RouteConfig{
		URI:    "/markov",
		Method: []string{http.MethodGet},
		Children: []router.RouteInitializerFunc{
			NewListRoute,
		},
	}
}

// swagger:route GET /api/markov markov markovGet
//
// Generate a markov chain based on a channel
//
//	Responses:
//		200: MarkovResponse
func (a *MarkovRoute) Handler() router.RouterHandler {
	return func(c *fiber.Ctx) (int, interface{}, error) {
		channel := c.Query("channel", "")
		seed := c.Query("seed", "")

		key := fmt.Sprintf("twitch:%s:chat-data", channel)
		u := c.Locals(locals.LocalRequestID).(uuid.UUID)

		if channel == "" {
			return http.StatusBadRequest, nil, fmt.Errorf("missing channel parameter")
		}

		storedData, err := a.Ctx.Inst().Redis.GetAllList(c.Context(), key)
		if err != nil {
			if err != redis.Nil {
				zap.S().Errorf("Failed to get channel data from redis: %s", err)

				return http.StatusInternalServerError, nil, router.ErrInternalServerError
			}

			req := pb.SubChannelReq{
				Channel: channel,
			}

			reqByte, err := proto.Marshal(&req)
			if err != nil {
				zap.S().Errorw("Failed to marshal protobuf message", "error", err)

				return http.StatusInternalServerError, nil, router.ErrInternalServerError
			}

			err = a.Ctx.Inst().RMQ.Publish(a.Ctx, rabbitmq.PublishSettings{
				RoutingKey: rabbitmq.QueueJoinRequest,
				Msg: amqp091.Publishing{
					Body:        reqByte,
					ContentType: "application/protobuf; twitch.SubChannelReq",
				},
			})

			if err != nil {
				zap.S().Errorw("Failed to send subcribe to RabbitMQ", "error", err)

				return http.StatusInternalServerError, nil, fmt.Errorf("chat logger is not available")
			}

			if err := a.Ctx.Inst().Redis.Publish(c.Context(), "twitch:chat-logger:join", reqByte); err != nil {
				zap.S().Errorw("Failed to publish to redis", "error", err)

				return http.StatusInternalServerError, nil, fmt.Errorf("chat logger is not available")
			}

			return http.StatusNotFound, nil, ErrNoData
		} else if l := len(storedData); l < 100 {

			return http.StatusInternalServerError, nil, ErrNotEnoughData(l)
		} else if !a.isAlive {
			return http.StatusInternalServerError, nil, ErrMarkovNotAlive
		}

		markovChan, err := a.genMarkov(c.Context(), u, storedData, seed)

		if err != nil {
			zap.S().Errorw("Failed to generate markov chain", "error", err)

			return http.StatusInternalServerError, nil, router.ErrInternalServerError
		}

		result, ok := <-markovChan

		if !ok {
			zap.S().Errorw("Took to long to generate markov chain", "channel", channel)

			return http.StatusInternalServerError, nil, ErrTooLong
		}

		if result.Error != nil {

			if strings.HasPrefix(*result.Error, "Failed to build a sentence after") {
				return http.StatusNotFound, nil, ErrUnableToGenerate
			} else {
				return http.StatusInternalServerError, nil, router.ErrInternalServerError
			}
		}

		return http.StatusOK, MarkovResponse{Markov: result.Result}, nil
	}
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
		case m := <-msg:
			{
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
		case <-a.Ctx.Done():
			return
		}
	}
}

func (a *MarkovRoute) genMarkov(ctx context.Context, corrId uuid.UUID, data []string, seed string) (chan *pb.MarkovResponse, error) {
	p := pb.MarkovRequest{
		Messages: data,
	}

	if seed != "" {
		p.Seed = &seed
	}

	reqByte, err := proto.Marshal(&p)
	if err != nil {
		return nil, err
	}

	_ = a.Ctx.Inst().RMQ.Publish(ctx, rabbitmq.PublishSettings{
		RoutingKey: rabbitmq.QueueMarkovGenenerator,
		Msg: amqp091.Publishing{
			CorrelationId: corrId.String(),
			Body:          reqByte,
		},
	})

	markovChan := make(chan *pb.MarkovResponse)

	go func() {
		if a.markovReqs[corrId.String()] == nil {
			a.markovReqs[corrId.String()] = make(chan *pb.MarkovResponse)
		}

		for {
			select {
			case <-ctx.Done():
				return
			case res, ok := <-a.markovReqs[corrId.String()]:
				{
					if !ok {
						return
					}
					markovChan <- res
					close(markovChan)
					delete(a.markovReqs, corrId.String())
					return
				}
			case <-time.After(10 * time.Second):
				{
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
