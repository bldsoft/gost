package alert

import (
	"context"
	"time"

	"github.com/bldsoft/gost/log"
	heap "github.com/bldsoft/gost/utils/heap"
	wp "github.com/bldsoft/gost/utils/worker_pool"
)

type ProcessorCfg struct {
	CheckInterval time.Duration
}

type Processor struct {
	Cfg ProcessorCfg
	Checker
	AlertHandler
}

type scheduledAlertProcessor struct {
	Processor
	Next time.Time
}

type AlertManager struct {
	alertProcessorsQueue *heap.AsyncHeap[scheduledAlertProcessor]
}

func NewAlertManager() *AlertManager {
	return &AlertManager{
		alertProcessorsQueue: heap.NewAsyncHeap(func(a, b scheduledAlertProcessor) bool {
			return a.Next.Before(b.Next)
		}),
	}
}

func (m *AlertManager) Run(ctx context.Context) error {
	var wp wp.WorkerPool
	wp.WithTaskChannel(make(chan func()))
	wp.SetWorkerN(1)
	defer wp.CloseAndWait()

	for m.alertProcessorsQueue.Len() > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		alertProcessor := m.alertProcessorsQueue.Pop()

		if err := m.waitUntil(ctx, alertProcessor.Next); err != nil {
			return err
		}

		wp.In() <- func() {
			var alerts []Alert
			alerts, err := alertProcessor.CheckAlerts(ctx)
			if err != nil {
				log.FromContext(ctx).ErrorfWithFields(log.Fields{
					"alert_processor": alertProcessor.Processor,
				}, "Failed to check alerts: %v", err)
				return
			}

			alertProcessor.Handle(ctx, alerts...)
		}

		alertProcessor.Next = time.Now().Add(alertProcessor.Cfg.CheckInterval)
		m.alertProcessorsQueue.Push(alertProcessor)
	}
	return nil
}

func (m *AlertManager) AddAlertProcessor(processor ...Processor) {
	for _, p := range processor {
		m.alertProcessorsQueue.Push(scheduledAlertProcessor{
			Processor: p,
		})
	}
}

func (m *AlertManager) waitUntil(ctx context.Context, notBefore time.Time) error {
	if notBefore.IsZero() {
		return ctx.Err()
	}

	if now := time.Now(); notBefore.After(now) {
		select {
		case <-time.After(notBefore.Sub(now)):
		case <-ctx.Done():
		}
	}
	return ctx.Err()
}
