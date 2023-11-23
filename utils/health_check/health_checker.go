package health_check

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/bldsoft/gost/log"
)

var ErrServiceNotAvailable = errors.New("service is not available")

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type healthCheck struct {
	err      error
	lastRead time.Time
}

type HealthChecker struct {
	cachedHealthChecks                      sync.Map // url -> healthCheck
	urlToMtx                                sync.Map
	healthCheckTimeout, healthCheckInterval time.Duration
	httpClient                              HttpClient
}

func NewHealthChecker(healthCheckTimeout, healthCheckInterval time.Duration) *HealthChecker {
	return &HealthChecker{healthCheckTimeout: healthCheckTimeout, healthCheckInterval: healthCheckInterval}
}

func (hc *HealthChecker) HealthCheck(ctx context.Context, url string) error {
	check, ok := hc.cachedHealthCheck(url)
	if ok {
		// updating lastRead timestamp to continue running the background health check goroutine
		firstCheck := check
		for {
			newCheck := healthCheck{
				err:      check.err,
				lastRead: time.Now(),
			}
			if hc.cachedHealthChecks.CompareAndSwap(url, check, newCheck) {
				return check.err
			}

			check, ok = hc.cachedHealthCheck(url)
			if !ok { //  background goroutine is finished
				break
			}
			if check.lastRead.After(firstCheck.lastRead) { // someone else updated lastRead timestamp
				return check.err
			}
		}
	}

	// the first request for a long time. Starting the background goroutine
	mtxI, _ := hc.urlToMtx.LoadOrStore(url, &sync.Mutex{})
	mtx := mtxI.(*sync.Mutex)
	mtx.Lock()
	defer mtx.Unlock()

	check, ok = hc.cachedHealthCheck(url)
	if !ok {
		check = healthCheck{
			err:      hc.checkWithRetries(context.Background(), url, hc.errRetry()),
			lastRead: time.Now(),
		}
		hc.cachedHealthChecks.Store(url, check)
		go hc.runHealthCheck(context.Background(), url, hc.HealthCheckTimeout(), hc.HealthCheckInterval())
	}

	hc.urlToMtx.CompareAndDelete(url, mtxI)
	return check.err
}

func (hc *HealthChecker) runHealthCheck(ctx context.Context, url string, healthCheckTimeout, updateInterval time.Duration) {
	t := time.NewTicker(updateInterval)
	defer t.Stop()

	log.FromContext(ctx).DebugWithFields(log.Fields{"URL": url}, "Health checker: started")
	for range t.C {
		check, _ := hc.cachedHealthCheck(url)
		if time.Since(check.lastRead) > hc.HealthCheckTTLWithoutRead() {
			if hc.cachedHealthChecks.CompareAndDelete(url, check) {
				log.FromContext(ctx).DebugWithFields(log.Fields{"URL": url}, "Health checker: finished")
				return
			}
			check, _ = hc.cachedHealthCheck(url)
		}
		if check.err == nil {
			check.err = hc.checkWithRetries(ctx, url, hc.errRetry())
		} else {
			check.err = hc.checkWithRetries(ctx, url, 1)
		}
		hc.cachedHealthChecks.Store(url, check)
		// log.FromContext(ctx).DebugOrErrorWithFields(check.err, log.Fields{"URL": url}, "Health checker")
	}
}

func (hc *HealthChecker) checkWithRetries(ctx context.Context, url string, retry int) error {
	var err error
	for i := 0; i < hc.errRetry(); i++ {
		log.Logger.WithFuncDuration(func() {
			err = hc.check(ctx, url)
		}).DebugOrErrorfWithFields(err, log.Fields{"URL": url}, "Health checker: ping (%d)", i+1)
		if err == nil {
			break
		}
	}
	return err
}

func (hc *HealthChecker) check(ctx context.Context, url string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create the request for the health check : %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, hc.HealthCheckTimeout())
	defer cancel()
	req = req.WithContext(ctx)

	res, err := hc.client().Do(req)
	if err != nil {
		return fmt.Errorf("failed to make the request for the health check : %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		return fmt.Errorf("%s: %w", url, ErrServiceNotAvailable)
	}

	return nil
}

func (hc *HealthChecker) cachedHealthCheck(url string) (check healthCheck, ok bool) {
	checkI, ok := hc.cachedHealthChecks.Load(url)
	if ok {
		return checkI.(healthCheck), true
	}
	return healthCheck{}, false
}

func (hc *HealthChecker) HealthCheckTimeout() time.Duration {
	if hc.healthCheckTimeout == 0 {
		return time.Second
	}
	return hc.healthCheckTimeout
}

func (hc *HealthChecker) HealthCheckTTLWithoutRead() time.Duration {
	return hc.HealthCheckInterval() * 2
}

func (hc *HealthChecker) HealthCheckInterval() time.Duration {
	if hc.healthCheckInterval == 0 {
		return 30 * time.Second
	}
	return hc.healthCheckInterval
}

func (hc *HealthChecker) errRetry() int {
	return 3
}

func (hc *HealthChecker) SetClient(client HttpClient) {
	hc.httpClient = client
}

func (hc *HealthChecker) client() HttpClient {
	if hc.httpClient == nil {
		return http.DefaultClient
	}
	return hc.httpClient
}
