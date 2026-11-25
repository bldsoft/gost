package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/bldsoft/gost/config"
	"github.com/bldsoft/gost/log"
)

// InternalServer serves platform-facing endpoints (probes, metrics) on a
// private address with its own router. It never touches http.DefaultServeMux,
// which Server owns, so any number of them can run next to a Server.
type InternalServer struct {
	addr   config.Address
	router chi.Router
	srv    *http.Server

	mu       sync.Mutex
	listener net.Listener
}

func NewInternalServer(addr config.Address) *InternalServer {
	router := chi.NewMux()

	return &InternalServer{
		addr:   addr,
		router: router,
		srv: &http.Server{
			Handler:           router,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
	}
}

func (s *InternalServer) Handle(pattern string, h http.Handler) *InternalServer {
	s.router.Handle(pattern, h)

	return s
}

func (s *InternalServer) Mount(pattern string, h http.Handler) *InternalServer {
	s.router.Mount(pattern, h)

	return s
}

// Addr is nil until Listen or Run has bound the listener.
func (s *InternalServer) Addr() net.Addr {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listener == nil {
		return nil
	}

	return s.listener.Addr()
}

// Listen binds without serving, so a bind error surfaces before Run is started in a goroutine.
func (s *InternalServer) Listen() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listener != nil {
		return nil
	}

	listener, err := net.Listen("tcp", s.addr.HostPort())
	if err != nil {
		return err
	}
	s.listener = listener

	return nil
}

func (s *InternalServer) Run() error {
	if err := s.Listen(); err != nil {
		return err
	}

	s.mu.Lock()
	listener := s.listener
	s.mu.Unlock()

	log.Infof("Internal server listening %s", listener.Addr().String())
	if err := s.srv.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *InternalServer) Stop(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
