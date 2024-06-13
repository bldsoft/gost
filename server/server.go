package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bldsoft/gost/log"
	gost_middleware "github.com/bldsoft/gost/server/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hashicorp/go-multierror"
)

type Router = chi.Router

var DefaultLogger func(next http.Handler) http.Handler = log.DefaultRequestLogger()

func defaultMiddlewares() chi.Middlewares {
	return chi.Middlewares{
		middleware.RequestID,
		gost_middleware.RealIP,
		middleware.Logger,
		middleware.Recoverer,
	}
}

type IMicroservice interface {
	BuildRoutes(Router)
	GetAsyncRunners() []AsyncRunner
}

type Server struct {
	srv               *http.Server
	router            *chi.Mux
	microservices     []IMicroservice
	commonMiddlewares chi.Middlewares
	runnerManager     *AsyncJobManager
	routerWrapper     func(http.Handler) http.Handler
}

func NewServer(config Config, microservices ...IMicroservice) *Server {
	srv := Server{srv: &http.Server{
		Addr:    config.ServiceBindAddress.HostPort(),
		Handler: nil},
		router:            chi.NewRouter(),
		microservices:     microservices,
		commonMiddlewares: nil,
		runnerManager:     NewAsyncJobManager()}
	middleware.DefaultLogger = DefaultLogger
	return &srv
}

func (s *Server) WithWriteTimeout(t time.Duration) *Server {
	s.srv.WriteTimeout = t
	return s
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

func (s *Server) SetRouterWrapper(middleware func(http.Handler) http.Handler) *Server {
	s.routerWrapper = middleware
	return s
}

func (s *Server) AddAsyncRunners(runners ...AsyncRunner) *Server {
	s.runnerManager.Append(runners...)
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

	if s.routerWrapper != nil {
		http.Handle("/", s.routerWrapper(s.router))
	} else {
		http.Handle("/", s.router)
	}
}

func (s *Server) Start() {
	s.init()

	go s.runnerManager.Start()
	go func() {
		log.Infof("Server started. Listening on %s", s.srv.Addr)
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

	errors := make([]error, 0)
	if err := s.srv.Shutdown(timeout); err != nil {
		errors = append(errors, err)
	}
	if err := s.runnerManager.Stop(timeout); err != nil {
		if merr, ok := err.(*multierror.Error); ok {
			errors = append(errors, merr.Errors...)
		} else {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		log.Errorf("Failed to shut down server gracefully")
		for _, err := range errors {
			log.Errorf("%v", err)
		}
	} else {
		log.Info("Server gracefully stopped")
	}
}
