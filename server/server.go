package server

import (
	"context"
	"net"
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
	httpListener      net.Listener
	httpsListener     net.Listener
	router            *chi.Mux
	microservices     []IMicroservice
	commonMiddlewares chi.Middlewares
	runnerManager     *AsyncJobManager
	routerWrapper     func(http.Handler) http.Handler
	config            Config
}

func NewServer(config Config, microservices ...IMicroservice) *Server {
	httpListener, err := net.Listen("tcp", config.ServiceBindAddressHTTP.HostPort())
	if err != nil {
		log.ErrorfWithFields(log.Fields{"err": err}, "Failed to create http listener on %s.", config.ServiceBindAddressHTTP.HostPort())
	}

	httpsListener, err := net.Listen("tcp", config.ServiceBindAddressHTTPS.HostPort())
	if err != nil {
		log.ErrorfWithFields(log.Fields{"err": err}, "Failed to create https listener on %s", config.ServiceBindAddressHTTPS.HostPort())
	}

	srv := Server{
		httpListener:      httpListener,
		httpsListener:     httpsListener,
		router:            chi.NewRouter(),
		microservices:     microservices,
		commonMiddlewares: nil,
		runnerManager:     NewAsyncJobManager(),
		config:            config,
	}
	middleware.DefaultLogger = DefaultLogger
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

	if s.httpListener != nil {
		log.Infof("Server listening http on %s", s.config.ServiceBindAddressHTTP.HostPort())
		go func() {
			if err := http.Serve(s.httpListener, nil); err != http.ErrServerClosed {
				log.Error(err.Error())
			}
		}()
	}

	if s.httpsListener != nil {
		log.Infof("Server listening https on %s", s.config.ServiceBindAddressHTTPS.HostPort())
		go func() {
			if err := http.ServeTLS(s.httpsListener, nil, s.config.TLSCertificatePath, s.config.TLSKeyPath); err != http.ErrServerClosed {
				log.Error(err.Error())
			}
		}()
	}

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
	if s.httpListener != nil {
		if err := s.httpListener.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	if s.httpsListener != nil {
		if err := s.httpsListener.Close(); err != nil {
			errors = append(errors, err)
		}
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
