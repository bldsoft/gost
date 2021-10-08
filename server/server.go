package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bldsoft/gost/log"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func defaultMiddlewares() chi.Middlewares {
	return chi.Middlewares{
		middleware.RequestID,
		middleware.RealIP,
		middleware.Logger,
		middleware.Recoverer,
	}
}

type IMicroservice interface {
	BuildRoutes(chi.Router)
	GetAsyncRunners() []AsyncRunner
}

type Server struct {
	srv               *http.Server
	router            *chi.Mux
	microservices     []IMicroservice
	commonMiddlewares chi.Middlewares
	runnerManager     *AsyncRunnerManager
}

func NewServer(config Config, microservices ...IMicroservice) *Server {
	srv := Server{srv: &http.Server{
		Addr:    config.ServiceAddress(),
		Handler: nil},
		router:            chi.NewRouter(),
		microservices:     microservices,
		commonMiddlewares: nil,
		runnerManager:     NewRunnerManager()}
	return &srv
}

func (s *Server) UseDefaultMiddlewares() *Server {
	s.commonMiddlewares = defaultMiddlewares()
	return s
}

func (s *Server) AppendMiddlewares(middlewares ...func(http.Handler) http.Handler) *Server {
	s.commonMiddlewares = append(s.commonMiddlewares, middlewares...)
	return s
}

func (s *Server) SetMiddlewares(middlewares ...func(http.Handler) http.Handler) *Server {
	s.commonMiddlewares = middlewares
	return s
}

func (s *Server) init() {
	for _, microservice := range s.microservices {
		s.runnerManager.Append(microservice.GetAsyncRunners()...)
	}

	s.router.Use(s.commonMiddlewares...)

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

	go func() {
		s.runnerManager.Start()
		log.Infof("Server started. Listening on %s\n", s.srv.Addr)
		if err := s.srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Error(err.Error())
		}
	}()
	s.gracefulShutdown()
}

func (s *Server) gracefulShutdown() {
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	defer stop()

	<-ctx.Done()

	timeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	errors := make([]error, 0, len(s.runnerManager.runners)+1)
	if err := s.srv.Shutdown(timeout); err != nil {
		errors = append(errors, err)
	}
	for err := range s.runnerManager.Stop(timeout) {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		log.ErrorfWithErrs(errors, "Failed to shut down server gracefully")
	} else {
		log.Info("Server gracefully stopped")
	}
}
