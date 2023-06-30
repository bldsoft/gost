package health_check

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bldsoft/gost/utils/health_check"
	"github.com/stretchr/testify/require"
)

type clientWithCounter struct {
	counter atomic.Int64
}

func (c *clientWithCounter) Do(req *http.Request) (*http.Response, error) {
	c.counter.Add(1)
	return http.DefaultClient.Do(req)
}

func TestHealthCheckerNoExtraChecks(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	}))
	defer svr.Close()

	healthCheckInterval := 10 * time.Millisecond
	healthChecker := health_check.NewHealthChecker(time.Second, healthCheckInterval)
	var client clientWithCounter
	healthChecker.SetClient(&client)

	goN := 10
	var wg sync.WaitGroup
	start, stop := make(chan struct{}), make(chan struct{})
	for i := 0; i < goN; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for {
				select {
				case <-stop:
					return
				default:
					require.NoError(t, healthChecker.HealthCheck(context.Background(), svr.URL))
				}
			}
		}()
	}

	waitTime := time.Second
	close(start)
	time.Sleep(waitTime)
	close(stop)
	wg.Wait()

	require.LessOrEqual(t, client.counter.Load(), waitTime/healthCheckInterval+1)
}
