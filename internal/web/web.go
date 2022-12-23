//		Version: 0.0.1
//
//		Consumes:
//		  - application/json
//
//		Produces:
//		  - application/json
//
// swagger:meta

package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/JoachimFlottorp/magnolia/internal/ctx"
	"github.com/JoachimFlottorp/magnolia/internal/mongo"
	"github.com/JoachimFlottorp/magnolia/internal/web/locals"
	"github.com/JoachimFlottorp/magnolia/internal/web/response"
	"github.com/JoachimFlottorp/magnolia/internal/web/router"
	"github.com/JoachimFlottorp/magnolia/internal/web/routes/api"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.uber.org/zap"
)

type Server struct {
	gCtx ctx.Context
	App  *fiber.App
}

func New(gCtx ctx.Context) error {
	port := gCtx.Config().Http.Port
	addr := fmt.Sprintf("%s:%d", "0.0.0.0", uint16(port))

	s := Server{
		gCtx: gCtx,
		App: fiber.New(fiber.Config{
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				if strings.HasPrefix(c.Path(), "/api") {

					res, _ := json.Marshal(&response.ApiResponse{
						Success:   false,
						RequestID: uuid.New(),
						Timestamp: time.Now(),
						Error:     "No such endpoint",
					})

					c.Set("Content-Type", "application/json")
					c.Status(http.StatusNotFound)
					return c.Send(res)
				}

				c.Status(http.StatusNotFound)
				c.Set("Content-Type", "text/html")
				return c.SendString("404 - Not found")
			},
		}),
	}

	s.App.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile("web/public/index.html")
	})

	s.App.Static("/public", "web/public")

	s.setupRoutes(api.NewApi(gCtx), s.App, "")

	go func() {
		<-gCtx.Done()

		_ = s.App.Shutdown()
	}()

	zap.S().Infof("Starting server on %s", addr)

	return s.App.Listen(addr)
}

func (s *Server) setupRoutes(r router.Route, parent fiber.Router, parentName string) {
	routeConfig := r.Configure()

	routeGroup := parent.Group(routeConfig.URI)

	handlers := []fiber.Handler{}

	handlers = append(handlers, s.beforeHandler())

	for _, middleware := range routeConfig.Middleware {
		handlers = append(handlers, middleware.Handler())
	}
	handlers = append(handlers, r.Handler())

	handlers = append(handlers, s.afterHandler())

	for _, method := range routeConfig.Method {
		switch method {
		case http.MethodGet:
			{
				routeGroup.Get("", handlers...)
				routeGroup.Get("/", handlers...)
			}
		case http.MethodPost:
			{
				routeGroup.Post("", handlers...)
				routeGroup.Post("/", handlers...)
			}
		default:
			{
				zap.S().Errorf("Unknown method %s", method)
			}
		}

		zap.S().Infof("Registered route %s %s", method, parentName+routeConfig.URI)
	}

	for _, routeChildren := range routeConfig.Children {
		childrenRouteConfig := routeChildren(s.gCtx)

		s.setupRoutes(childrenRouteConfig, routeGroup, parentName+routeConfig.URI)
	}
}

func (s *Server) beforeHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals(locals.LocalRequestID, uuid.New())

		return c.Next()
	}
}

func (s *Server) afterHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		t := time.Now()

		if c.Locals(locals.LocalStatus) == nil {
			c.Locals(locals.LocalStatus, http.StatusOK)
		}

		statusCode := c.Locals(locals.LocalStatus).(int)
		body := c.Locals(locals.LocalResponse)

		apiRes := response.ApiResponse{
			RequestID: c.Locals(locals.LocalRequestID).(uuid.UUID),
			Timestamp: t,
		}

		if statusCode != http.StatusOK {
			apiRes.Success = false

			err := c.Locals(locals.LocalError)
			if err != nil {
				apiRes.Error = fmt.Sprintf("%v", err)
			} else {
				apiRes.Error = http.StatusText(statusCode)
			}
		} else {
			apiRes.Success = true
			data, err := json.Marshal(body)
			if err != nil {
				zap.S().Errorf("Error marshaling response: %v", err)

				apiRes.Success = false
				apiRes.Error = http.StatusText(http.StatusInternalServerError)
			}

			apiRes.Data = data
		}

		log := mongo.ApiLog{
			ID:        primitive.NewObjectID(),
			Timestamp: t,
			Method:    c.Method(),
			URL:       c.OriginalURL(),
			Status:    fmt.Sprintf("%v", c.Locals(locals.LocalStatus)),
			IP:        c.Get("X-Forwarded-For", "?"),
			UserAgent: c.Get("User-Agent", "?"),
			Body:      string(apiRes.Data),
		}

		if !apiRes.Success {
			log.Error = apiRes.Error
		}

		s.gCtx.Inst().Mongo.Collection(mongo.CollectionAPILog).InsertOne(s.gCtx, log)

		c.Status(statusCode)
		return c.JSON(apiRes)
	}
}
