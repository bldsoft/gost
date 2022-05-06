package stat

import (
	"context"
	"sync"
	"time"
)

type StatCollector interface {
	Stat(ctx context.Context) Stat
}

type Service struct {
	statCollectors []StatCollector
}

func NewService(statCollectors ...StatCollector) *Service {
	return &Service{statCollectors: statCollectors}
}

func (s *Service) AddStatCollectors(statCollectors ...StatCollector) {
	s.statCollectors = append(s.statCollectors, statCollectors...)
}

func (s *Service) Stats(ctx context.Context) []Stat {
	var wg sync.WaitGroup
	wg.Add(len(s.statCollectors))

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statC := make(chan Stat, len(s.statCollectors))
	for _, sc := range s.statCollectors {
		go func(ctx context.Context, collector StatCollector) {
			defer wg.Done()
			statC <- collector.Stat(ctx)
		}(ctx, sc)
	}
	wg.Wait()
	close(statC)

	stats := make([]Stat, 0, len(s.statCollectors))
	for stat := range statC {
		stats = append(stats, stat)
	}
	return stats
}
