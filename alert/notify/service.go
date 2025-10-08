package notify

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bldsoft/gost/alert/notify/channel"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/utils"
	"github.com/bldsoft/gost/utils/poly"
	wp "github.com/bldsoft/gost/utils/worker_pool"
)

type Notification struct {
	Receiver poly.Poly[Receiver]
	Message  channel.Message
}

type RetriedNotification struct {
	Notification Notification
	RetryAt      time.Time
	RetryCount   int
}

type Queue interface {
	Enqueue(ctx context.Context, n RetriedNotification) error
	Dequeue(ctx context.Context) (id string, n *RetriedNotification, err error)
	MarkDone(ctx context.Context, id string) error
	Requeue(ctx context.Context, id string, n RetriedNotification) error
}

type ServiceConfig struct {
	RetryCount             int
	RetryQueuePollInterval time.Duration
	WorkerN                int
	SendTimeout            time.Duration

	Dispatcher DispatcherConfig
}

var DefaultNotificationServiceConfig = ServiceConfig{
	RetryCount:             1,
	RetryQueuePollInterval: 1 * time.Minute,
	WorkerN:                1,
	SendTimeout:            10 * time.Second,
}

type Service struct {
	cfg        ServiceConfig
	dispatcher *Dispatcher
	queue      Queue
	wg         *wp.WorkerPool
}

func NewService(cfg ServiceConfig) *Service {
	res := &Service{
		cfg:        cfg,
		dispatcher: NewDispatcher(cfg.Dispatcher),
		queue:      nil,
		wg:         new(wp.WorkerPool).SetWorkerN(int64(cfg.WorkerN)),
	}
	if cfg.RetryCount > 0 {
		res.queue = NewMemoryQueue(1024)
	}
	return res
}

func (ns *Service) SetQueue(queue Queue) *Service {
	ns.queue = queue
	return ns
}

func (ns *Service) withSendTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	timeout := cmp.Or(ns.cfg.SendTimeout, DefaultNotificationServiceConfig.SendTimeout)
	return context.WithTimeout(ctx, timeout)
}

func (ns *Service) Send(ctx context.Context, notification Notification) error {
	ctx, cancel := ns.withSendTimeout(ctx)
	defer cancel()

	if err := ns.dispatcher.Send(ctx, notification); err != nil {
		if ns.queue == nil {
			return err
		}

		log.FromContext(ctx).ErrorfWithFields(log.Fields{
			"notification": notification,
			"error":        err,
		}, "Failed to send notification")

		ns.queue.Enqueue(ctx, RetriedNotification{
			Notification: notification,
			RetryAt:      time.Now().Add(ns.cfg.RetryQueuePollInterval),
			RetryCount:   ns.cfg.RetryCount,
		})
	}
	return nil
}

func (ns *Service) Run(ctx context.Context) error {
	ticker := time.NewTicker(ns.cfg.RetryQueuePollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			for {
				id, n, err := ns.queue.Dequeue(ctx)
				if err != nil {
					if !errors.Is(err, utils.ErrObjectNotFound) {
						log.FromContext(ctx).ErrorfWithFields(log.Fields{
							"error": err,
						}, "Failed to dequeue retried notification")
					}
					break
				}
				if n == nil {
					// empty queue
					break
				}
				ns.wg.In() <- func() {
					if err := ns.retrySend(ctx, id, *n); err != nil {
						log.FromContext(ctx).ErrorfWithFields(log.Fields{
							"id":           id,
							"notification": n.Notification,
							"error":        err,
						}, "Failed to retry send")
					}
				}

			}
		}
	}
}

func (ns *Service) retrySend(ctx context.Context, id string, n RetriedNotification) error {
	ctx, cancel := ns.withSendTimeout(ctx)
	defer cancel()

	sendErr := ns.dispatcher.Send(ctx, n.Notification)
	if sendErr == nil {
		if err := ns.queue.MarkDone(ctx, id); err != nil {
			return fmt.Errorf("mark done after success: %w", err)
		}
		return nil
	}

	n.RetryCount--
	if n.RetryCount <= 0 {
		if err := ns.queue.MarkDone(ctx, id); err != nil {
			return errors.Join(sendErr, fmt.Errorf("mark done after failure: %w", err))
		}
		return sendErr
	}

	n.RetryAt = time.Now().Add(ns.cfg.RetryQueuePollInterval)
	if err := ns.queue.Requeue(ctx, id, n); err != nil {
		return errors.Join(sendErr, fmt.Errorf("requeue: %w", err))
	}

	return sendErr
}
