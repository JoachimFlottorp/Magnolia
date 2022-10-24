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
	"log"
	"net"
	"net/http"
	"time"

	"github.com/JoachimFlottorp/magnolia/internal/ctx"
	"github.com/JoachimFlottorp/magnolia/internal/mongo"
	"github.com/JoachimFlottorp/magnolia/internal/web/response"
	"github.com/JoachimFlottorp/magnolia/internal/web/router"
	"github.com/JoachimFlottorp/magnolia/internal/web/routes/api"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type HandlerFunc func(w http.ResponseWriter, r *http.Request) response.RouterResponse

type l struct {
	logger *zap.SugaredLogger
}

func (log *l) Write(p []byte) (int, error) {
	log.logger.Errorw("HTTPError", "error", string(p))
	return len(p), nil
}

type Server struct {
	gCtx 		ctx.Context
	listener 	net.Listener
	router 		*mux.Router
}

func New(gCtx ctx.Context) error {
	port := gCtx.Config().Http.Port
	addr := fmt.Sprintf("%s:%d", "0.0.0.0", port)

	s := Server{
		gCtx: gCtx,
	}

	var err error
	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	s.router = mux.NewRouter().StrictSlash(false)

	logger := log.New(&l{zap.S()}, "", 0)

	server := http.Server{
		Handler: s.router,
		ErrorLog: logger,
		ReadTimeout: 20 * time.Second,
		WriteTimeout: 20 * time.Second,
	}

	s.router.Path("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f := "web/public/index.html"
		http.ServeFile(w, r, f)
	}))

	s.router.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("web/public"))))

	s.setupRoutes(api.NewApi(gCtx), s.router)

	s.router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path[:4] == "/api" {
			res, _ := json.Marshal(&response.ApiResponse{
				Success: false,
				RequestID: uuid.New(),
				Timestamp: time.Now(),
				Error: http.StatusText(http.StatusNotFound),
			})

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write(res)
			return
		} 
		
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - Not found"))
	})
	
	go func() {
		<-gCtx.Done()

		_ = server.Shutdown(gCtx)
	}()

	go func() {
		time.Sleep(1 * time.Second)
		select {
		case <-gCtx.Done():
			return
		default:
			zap.S().Infof("Listening on %s", addr)
		}
	}()

	return server.Serve(s.listener)
}

func (s *Server) setupRoutes(r router.Route, parent *mux.Router) {
	routeConfig := r.Configure()

	route := parent.
		PathPrefix(routeConfig.URI).
		Methods(routeConfig.Method...).
		Subrouter().
		StrictSlash(false)

	// Allow endpoint without trailing slash
	route.HandleFunc("", s.wrapRouterHandler(r.Handler))
	route.HandleFunc("/", s.wrapRouterHandler(r.Handler))

	zap.S().
		With("route", routeConfig.URI,).
		Debug("Setup route")

	for _, child := range routeConfig.Children {
		s.setupRoutes(child, route)
	}

	for _, middleware := range routeConfig.Middleware {
		route.Use(middleware)
	}
}

func (s *Server) wrapRouterHandler(fn HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		res := fn(w, r)
		u := uuid.New()
		t := time.Now()

		for k, v := range res.Headers {
			w.Header().Set(k, v)
		}

		apiRes := response.ApiResponse {
			RequestID: u,
			Timestamp: t,
		}

		if res.Error != nil {
			apiRes.Error = res.Error.Error()
			apiRes.Success = false
		} else {
			apiRes.Success = true
			apiRes.Data = res.Body
		}
		
		go func() {
			log := mongo.ApiLog {
				ID: primitive.NewObjectID(),
				Timestamp: t,
				Method: r.Method,
				Path: r.URL.Path,
				Status: res.StatusCode,
				IP: fmt.Sprintf("%s (%s)", r.Header.Get("X-Forwarded-For"), r.RemoteAddr),
				UserAgent: r.UserAgent(),
			}

			if !apiRes.Success {
				log.Error = apiRes.Error
			}
	
			s.gCtx.Inst().Mongo.Collection(mongo.CollectionAPILog).InsertOne(s.gCtx, log)
		}()
		
		j, err := json.MarshalIndent(apiRes, "", "  ")

		if err != nil {
			zap.S().Errorw("Failed to marshal response", "error", err)
			
			j, _ := json.MarshalIndent(&response.ApiResponse{
				Success: false,
				RequestID: uuid.New(),
				Timestamp: time.Now(),
				Error: "Internal server error",
			}, "", "  ")
			
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(j)
			return
		}

		w.WriteHeader(res.StatusCode)
		w.Write(j)
	}
}