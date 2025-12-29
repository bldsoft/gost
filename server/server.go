package server

import (
	"context"
	"errors"
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
)

type Router = chi.Router

type connContextKey struct{}

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

type InitializableMicroservice interface {
	IMicroservice
	Initialize()
}

type Server struct {
	httpListener      net.Listener
	httpsListener     net.Listener
	srv               *http.Server
	microservices     []IMicroservice
	commonMiddlewares chi.Middlewares
	runnerManager     *AsyncJobManager
	routerWrapper     func(http.Handler) http.Handler
	needHealthProbes  bool
	config            Config
}

func NewServer(config Config, microservices ...IMicroservice) *Server {
	var (
		httpListener, httpsListener net.Listener
		err                         error
	)

	httpListener, err = net.Listen("tcp", config.ServiceBindAddress.HostPort())
	if err != nil {
		log.ErrorfWithFields(log.Fields{"err": err}, "Failed to create http listener on %s.", config.ServiceBindAddress.HostPort())
	}

	if config.TLS.IsTLSEnabled() {
		httpsListener, err = net.Listen("tcp", config.TLS.ServiceBindAddress.HostPort())
		if err != nil {
			log.ErrorfWithFields(log.Fields{"err": err}, "Failed to create https listener on %s", config.TLS.ServiceBindAddress.HostPort())
		}
	}

	srv := Server{
		httpListener:  httpListener,
		httpsListener: httpsListener,
		srv: &http.Server{
			ConnContext: func(ctx context.Context, c net.Conn) context.Context {
				ctx = context.WithValue(ctx, connContextKey{}, c)
				return ctx
			},
			Handler: nil},
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

func (s *Server) WithHealthProbes(v bool) *Server {
	s.needHealthProbes = v
	return s
}

func (s *Server) AddAsyncRunners(runners ...AsyncRunner) *Server {
	s.runnerManager.Append(runners...)
	return s
}

func (s *Server) init() {
	if !s.needHealthProbes {
		http.Handle("/", s.appRouter())
		return
	}

	dynamicRouter := new(dynamicRouter)
	dynamicRouter.Set(s.probesOnlyRouter())
	http.Handle("/", dynamicRouter)
	go func() {
		dynamicRouter.Set(s.appRouter())
	}()
}

func (s *Server) appRouter() http.Handler {
	defer func() {
		for _, microservice := range s.microservices {
			s.runnerManager.Append(microservice.GetAsyncRunners()...)
		}
		go s.runnerManager.Start()
	}()

	appRouter := s.newRouter(true)
	for _, m := range s.microservices {
		appRouter.Group(func(r chi.Router) {
			if im, ok := m.(InitializableMicroservice); ok {
				im.Initialize()
			}
			m.BuildRoutes(r)
		})
	}

	if s.routerWrapper != nil {
		return s.routerWrapper(appRouter)
	}
	return appRouter
}

func (s *Server) probesOnlyRouter() http.Handler {
	return s.newRouter(false)
}

func (s *Server) newRouter(isAppRouter bool) chi.Router {
	r := chi.NewMux()
	r.Use(s.commonMiddlewares...)
	r.Mount("/debug", middleware.Profiler())
	if s.needHealthProbes {
		r.Route("/probes", newProbesController().SetReady(isAppRouter).Mount)
	}
	return r
}

func (s *Server) Start() {
	s.init()

	s.serveListener(s.httpListener, "", "")
	s.serveListener(s.httpsListener, s.config.TLS.CertificatePath, s.config.TLS.KeyPath)

	s.gracefulShutdown()
}

func (s *Server) serveListener(listener net.Listener, certPath, keyPath string) {
	if listener == nil {
		return
	}

	log.Infof("Server listening %s", listener.Addr().String())
	go func() {
		var err error
		if certPath != "" && keyPath != "" {
			err = s.srv.ServeTLS(listener, certPath, keyPath)
		} else {
			err = s.srv.Serve(listener)
		}
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error(err.Error())
		}
	}()
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

	errs := errors.Join(s.srv.Shutdown(ctx), s.runnerManager.Stop(timeout))

	if errs != nil {
		log.Errorf("Failed to shut down server gracefully\n%v", errs)
	} else {
		log.Info("Server gracefully stopped")
	}
}
