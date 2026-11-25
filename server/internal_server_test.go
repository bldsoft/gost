package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// blockingMicroservice holds Initialize until release is closed, which keeps the
// application router unbuilt for as long as a test needs the "not ready" state.
type blockingMicroservice struct {
	release chan struct{}
}

func newBlockingMicroservice() *blockingMicroservice {
	return &blockingMicroservice{release: make(chan struct{})}
}

func (m *blockingMicroservice) Initialize() {
	<-m.release
}

func (m *blockingMicroservice) BuildRoutes(r Router) {
	r.Get("/ping", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("pong"))
	})
}

func (m *blockingMicroservice) GetAsyncRunners() []AsyncRunner {
	return nil
}

func get(h http.Handler, target string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))

	return rec
}

func eventuallyStatus(t *testing.T, h http.Handler, target string, want int) {
	t.Helper()
	require.Eventually(t, func() bool {
		return get(h, target).Code == want
	}, 5*time.Second, 10*time.Millisecond)
}

func TestInternalServer_ServesOnItsOwnAddress(t *testing.T) {
	internal := NewInternalServer("127.0.0.1:0").
		Handle("/metrics", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("# metrics"))
		}))

	require.Nil(t, internal.Addr())
	require.NoError(t, internal.Listen())
	require.NotNil(t, internal.Addr())

	done := make(chan error, 1)
	go func() { done <- internal.Run() }()

	resp, err := http.Get("http://" + internal.Addr().String() + "/metrics")
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, resp.Body.Close())
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "# metrics", string(body))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	require.NoError(t, internal.Stop(ctx))
	require.NoError(t, <-done)
}

func TestServer_ProbesMoveToInternalServer(t *testing.T) {
	ms := newBlockingMicroservice()
	cfg := Config{ServiceName: "test", ProbesBindAddress: "127.0.0.1:0"}
	metrics := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("# metrics"))
	})

	internal := NewInternalServer(cfg.ProbesBindAddress).
		Handle("/metrics", metrics)
	srv := NewServer(cfg, ms).
		UseDefaultMiddlewares().
		WithHealthProbes(true, internal)

	main := srv.handler()

	require.Equal(t, http.StatusNotFound, get(main, "/probes/ready").Code, "probes must leave the main listener")
	require.Equal(t, http.StatusNotFound, get(main, "/metrics").Code)
	require.Equal(t, http.StatusOK, get(internal.router, "/metrics").Code)
	require.Equal(t, http.StatusOK, get(internal.router, "/probes/healthy").Code)
	require.Equal(t, http.StatusOK, get(internal.router, "/probes/startup").Code)
	require.Equal(t, http.StatusServiceUnavailable, get(internal.router, "/probes/ready").Code)

	close(ms.release)

	eventuallyStatus(t, internal.router, "/probes/ready", http.StatusOK)
	eventuallyStatus(t, main, "/ping", http.StatusOK)
	require.Equal(t, http.StatusNotFound, get(main, "/probes/ready").Code)
}

func TestServer_ProbesBindAddressCreatesInternalServer(t *testing.T) {
	ms := newBlockingMicroservice()
	close(ms.release)
	srv := NewServer(Config{ServiceName: "test", ProbesBindAddress: "127.0.0.1:0"}, ms).
		WithHealthProbes(true)

	main := srv.handler()

	require.NotNil(t, srv.internalSrv)
	require.Equal(t, http.StatusNotFound, get(main, "/probes/healthy").Code)
	require.Equal(t, http.StatusOK, get(srv.internalSrv.router, "/probes/healthy").Code)
	eventuallyStatus(t, srv.internalSrv.router, "/probes/ready", http.StatusOK)
}

func TestServer_ProbesStayOnMainListenerByDefault(t *testing.T) {
	ms := newBlockingMicroservice()
	srv := NewServer(Config{ServiceName: "test"}, ms).WithHealthProbes(true)

	main := srv.handler()

	require.Nil(t, srv.internalSrv)
	require.Equal(t, http.StatusOK, get(main, "/probes/healthy").Code)
	require.Equal(t, http.StatusServiceUnavailable, get(main, "/probes/ready").Code)

	close(ms.release)

	eventuallyStatus(t, main, "/probes/ready", http.StatusOK)
	require.Equal(t, http.StatusOK, get(main, "/ping").Code)
}

func TestServer_EmptyInternalAddressKeepsProbesOnMainListener(t *testing.T) {
	ms := newBlockingMicroservice()
	close(ms.release)
	srv := NewServer(Config{ServiceName: "test"}, ms).
		WithHealthProbes(true, NewInternalServer(""))

	main := srv.handler()

	require.Nil(t, srv.internalSrv)
	require.Equal(t, http.StatusOK, get(main, "/probes/healthy").Code)
	eventuallyStatus(t, main, "/probes/ready", http.StatusOK)
}

func TestServer_WithHealthProbesRejectsSeveralInternalServers(t *testing.T) {
	require.Panics(t, func() {
		NewServer(Config{ServiceName: "test"}).
			WithHealthProbes(true, NewInternalServer("127.0.0.1:0"), NewInternalServer("127.0.0.1:0"))
	})
}
