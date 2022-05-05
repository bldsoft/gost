package healthcheck

import (
	"context"
	"sync"
	"time"
)

type HealthChecker interface {
	CheckHealth(ctx context.Context) Health
}

type HealthCheckService struct {
	healthCheckers []HealthChecker
}

func NewHealthCheckService(healthCheckers ...HealthChecker) *HealthCheckService {
	return &HealthCheckService{healthCheckers: healthCheckers}
}

func (s *HealthCheckService) AddHealthCheckers(healthCheckers ...HealthChecker) {
	s.healthCheckers = append(s.healthCheckers, healthCheckers...)
}

func (s *HealthCheckService) CheckHealth(ctx context.Context) []Health {
	var wg sync.WaitGroup
	wg.Add(len(s.healthCheckers))

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	healthC := make(chan Health, len(s.healthCheckers))
	for _, hc := range s.healthCheckers {
		go func(ctx context.Context, checker HealthChecker) {
			defer wg.Done()
			healthC <- checker.CheckHealth(ctx)
		}(ctx, hc)
	}
	wg.Wait()
	close(healthC)

	healthChecks := make([]Health, 0, len(s.healthCheckers))
	for hc := range healthC {
		healthChecks = append(healthChecks, hc)
	}
	return healthChecks
}
