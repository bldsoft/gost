package alert

import (
	"context"
	"time"

	"github.com/bldsoft/gost/log"
	wp "github.com/bldsoft/gost/utils/worker_pool"
)

const failedCheckRetryInterval = 5 * time.Minute

type Processor struct {
	ID      string
	Source  Source
	Handler Handler
}

type Config struct {
	WorkerN int
}

type Manager struct {
	queue *queue[Processor]
	wp    *wp.WorkerPool
}

func NewManager(cfg Config) *Manager {
	return &Manager{
		queue: newQueue[Processor](),
		wp:    new(wp.WorkerPool).SetWorkerN(int64(cfg.WorkerN)),
	}
}

func (m *Manager) AddProcessor(p Processor) {
	m.queue.Push(p, time.Now())
}

func (m *Manager) RemoveProcessor(id string) {
	m.queue.RemoveFirstFunc(func(p Processor) bool {
		return p.ID == id
	})
}

func (m *Manager) Run(ctx context.Context) {
	for p := range m.queue.SyncSeq(ctx) {
		m.wp.In() <- func() {
			var next time.Time
			defer func() {
				if next.IsZero() {
					next = time.Now().Add(failedCheckRetryInterval)
				}
				m.queue.Push(p, next)
			}()
			defer func() {
				if err := recover(); err != nil {
					log.FromContext(ctx).ErrorfWithFields(log.Fields{
						"error": err,
					}, "Failed to process alerts")
				}
			}()

			alerts, next, err := p.Source.EvaluateAlerts(ctx)
			if err != nil {
				log.FromContext(ctx).ErrorfWithFields(log.Fields{
					"error": err,
				}, "Failed to evaluate alerts")
				return
			}

			p.Handler.Handle(ctx, alerts...)
		}
	}
	m.wp.CloseAndWait()
}
