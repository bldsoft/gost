package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bldsoft/gost/log"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type IMicroservice interface {
	BuildRoutes(*chi.Mux)
	Start()
	Stop(ctx context.Context)
}

type Server struct {
	srv          *http.Server
	router       *chi.Mux
	microservice IMicroservice
}

func NewServer(config Config, microservice IMicroservice) *Server {
	srv := Server{srv: &http.Server{Addr: config.ServiceAddress(), Handler: nil}, router: chi.NewRouter(), microservice: microservice}
	srv.init()
	return &srv
}

func (s *Server) init() {
	s.router.Use(middleware.Recoverer)

	s.microservice.BuildRoutes(s.router)
	s.router.Mount("/debug", middleware.Profiler())

	http.Handle("/", s.router)
}

func (s *Server) Start() {
	s.microservice.Start()
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
	s.microservice.Stop(ctxShutDown)
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
