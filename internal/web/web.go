//		Version: 0.0.1
//
//		Consumes:
//		- application/json
//
//		Produces:
//		- application/json
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

	"github.com/JoachimFlottorp/yeahapi/internal/ctx"
	"github.com/JoachimFlottorp/yeahapi/internal/web/router"
	"github.com/JoachimFlottorp/yeahapi/internal/web/routes/api"
	"github.com/google/uuid"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type l struct {
	logger *zap.SugaredLogger
}

func (log *l) Write(p []byte) (int, error) {
	log.logger.Errorw("HTTPError", "error", string(p))
	return len(p), nil
}

type Server struct {
	listener net.Listener
	router *mux.Router
}

func New(gCtx ctx.Context) error {
	port := gCtx.Config().Http.Port
	addr := fmt.Sprintf("%s:%d", "localhost", port)

	s := Server{}

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
			res, _ := json.Marshal(&router.ApiFail{
				Success: false,
				RequestID: uuid.New(),
				Timestamp: time.Now(),
				Error: "Not found",
			})

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write(res)
			return
		} 
		
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - Not found"))
		return
	})
	
	go func() {
		<-gCtx.Done()

		_ = server.Shutdown(gCtx)
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
	route.HandleFunc("", r.Handler)
	route.HandleFunc("/", r.Handler)

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