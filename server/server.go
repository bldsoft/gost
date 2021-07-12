package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/bldsoft/gost/log"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type IMicroservice interface {
	BuildRoutes(chi.Router)
	Start() error
	Stop(ctx context.Context) error
}

type Server struct {
	srv               *http.Server
	router            *chi.Mux
	microservices     []IMicroservice
	commonMiddlewares chi.Middlewares
}

func NewServer(config Config, microservices ...IMicroservice) *Server {
	srv := Server{srv: &http.Server{Addr: config.ServiceAddress(), Handler: nil}, router: chi.NewRouter(), microservices: microservices}
	return &srv
}

func (s *Server) DefaultMiddlewares() chi.Middlewares {
	return chi.Middlewares{
		middleware.RequestID,
		middleware.RealIP,
		middleware.Logger,
		middleware.Recoverer,
	}
}

func (s *Server) SetCommonMiddlewares(middlewares ...func(http.Handler) http.Handler) *Server {
	s.commonMiddlewares = middlewares
	return s
}

func (s *Server) init() {
	if s.commonMiddlewares == nil {
		s.router.Use(s.DefaultMiddlewares()...)
	} else {
		s.router.Use(s.commonMiddlewares...)
	}

	s.router.Mount("/debug", middleware.Profiler())

	for _, m := range s.microservices {
		s.router.Group(func(r chi.Router) {
			m.BuildRoutes(r)
		})
	}

	http.Handle("/", s.router)
}

func (s *Server) Start() {
	s.init()

	for _, m := range s.microservices {
		if err := m.Start(); err != nil {
			log.FatalWithFields(log.Fields{"microservice": m, "err": err}, "Failed to start microservice")
		}
	}

	go func() {
		log.Infof("Server started. Listening on %s\n", s.srv.Addr)
		if err := s.srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err.Error())
		}
	}()
	s.gracefulShutdown()
}

func (s *Server) stop() {
	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.srv.Shutdown(ctxShutDown); err != nil {
		log.Fatalf("Server shutdown failed:%+s", err)
	}

	var wg sync.WaitGroup
	wg.Add(len(s.microservices))
	for _, m := range s.microservices {
		go func(m IMicroservice) {
			defer wg.Done()
			if err := m.Stop(ctxShutDown); err != nil {
				log.ErrorWithFields(log.Fields{"microservice": m, "err": err}, "Failed to stop microservice")
			}
		}(m)
	}
	wg.Wait()
}

func (s *Server) gracefulShutdown() {
	stopped := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)

		signal.Notify(sigint, os.Interrupt)
		signal.Notify(sigint, syscall.SIGTERM)

		<-sigint

		s.stop()
		close(stopped)
	}()
	<-stopped
	log.Info("Server gracefully stopped")
}
