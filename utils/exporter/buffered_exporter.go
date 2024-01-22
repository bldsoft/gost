package exporter

import (
	"context"
	"fmt"
	"time"

	"github.com/bldsoft/gost/utils/ringbuf"
)

const (
	DefaultMaxBatchSize     = 1000
	DefaultMaxFlushInterval = time.Second
	DefaultChanBufSize      = 16384
)

type Fields = map[string]interface{}
type Logger interface {
	ErrorfWithFields(fields Fields, format string, v ...interface{})
	TracefWithFields(fields Fields, format string, v ...interface{})
}

type nullLogger struct{}

func (nullLogger) ErrorfWithFields(fields Fields, format string, v ...interface{}) {}
func (nullLogger) TracefWithFields(fields Fields, format string, v ...interface{}) {}

type BufferedExporterConfig struct {
	MaxFlushInterval time.Duration
	MaxBatchSize     int
	ChanBufSize      int  // buffer between writer and the goroutine that actually exports data
	PreserveOld      bool // false - discard old data in case of overflow, true - discard new

	Logger Logger
}

type BufferedExporter[T any] struct {
	exporter Exporter[T]
	cfg      BufferedExporterConfig

	writeC   chan T
	ringBuf  *ringbuf.RingBuf[T]
	flushBuf []T

	stop    chan struct{}
	stopped chan struct{}
}

func NewBuffered[T any](
	exporter Exporter[T],
	cfg BufferedExporterConfig,
) *BufferedExporter[T] {
	if cfg.ChanBufSize <= 0 {
		cfg.ChanBufSize = DefaultChanBufSize
	}
	if cfg.Logger == nil {
		cfg.Logger = nullLogger{}
	}
	return &BufferedExporter[T]{
		exporter: exporter,
		cfg:      cfg,
		writeC:   make(chan T, cfg.ChanBufSize),
		stop:     make(chan struct{}),
		stopped:  make(chan struct{}),
	}
}

func (be *BufferedExporter[T]) WithLogger(logger Logger) *BufferedExporter[T] {
	be.cfg.Logger = logger
	return be
}

func (be BufferedExporter[T]) MaxBatchSize() int {
	if be.cfg.MaxBatchSize > 0 {
		return be.cfg.MaxBatchSize
	}
	return DefaultMaxBatchSize
}

func (be BufferedExporter[T]) MaxFlushInterval() time.Duration {
	if be.cfg.MaxFlushInterval > 0 {
		return be.cfg.MaxFlushInterval
	}
	return DefaultMaxFlushInterval
}

func (be *BufferedExporter[T]) Export(items ...T) (n int, err error) {
	select {
	case <-be.stop:
		// do nothing
		return 0, nil
	default:
		return be.writeToChan(items...), nil
	}
}

func (be *BufferedExporter[T]) writeToChan(items ...T) (n int) {
	for i, item := range items {
		select {
		case be.writeC <- item:
		default:
			// discard: channel is full
			return i
		}
	}
	return len(items)
}

func (be *BufferedExporter[T]) flush() (err error) {
	defer func() {
		if e := recover(); e != nil {
			switch x := e.(type) {
			case error:
				err = x
			default:
				err = fmt.Errorf("%v", x)
			}
		}
	}()
	n := be.ringBuf.Copy(be.flushBuf)
	if _, err := be.exporter.Export(be.flushBuf[:n]...); err != nil {
		return err
	}
	be.ringBuf.Clear()
	return nil
}

func (be *BufferedExporter[T]) Run() error {
	defer close(be.stopped)

	be.ringBuf = ringbuf.New[T](be.MaxBatchSize()).WithOverwrite(!be.cfg.PreserveOld)
	be.flushBuf = make([]T, be.MaxBatchSize())

	ticker := time.NewTicker(be.MaxFlushInterval())
	defer ticker.Stop()

	lastFlushTime := time.Now()

	flush := func() {
		defer func() {
			lastFlushTime = time.Now()
		}()
		logFields := Fields{
			"current batch": be.ringBuf.Len(),
			"queued":        len(be.writeC),
		}
		if err := be.flush(); err != nil {
			logFields["err"] = err
			be.cfg.Logger.ErrorfWithFields(logFields, "buffered exporter")
			return
		}
		be.cfg.Logger.TracefWithFields(logFields, "buffered exporter")
	}

	for {
		select {
		case item := <-be.writeC:
			fullBefore := be.ringBuf.Full()
			_ = be.ringBuf.Push(item)
			if becameFull := !fullBefore && be.ringBuf.Full(); becameFull {
				flush()
			}
		case <-ticker.C:
			if !be.ringBuf.Empty() && time.Since(lastFlushTime) >= be.MaxFlushInterval() {
				flush()
			}
		case <-be.stop:
			close(be.writeC)

			for item := range be.writeC {
				if be.ringBuf.Full() {
					if err := be.flush(); err != nil {
						return err
					}
				}
				_ = be.ringBuf.Push(item)
			}
			return be.flush()
		}
	}
}

func (be *BufferedExporter[T]) Stop(ctx context.Context) error {
	close(be.stop)
	select {
	case <-be.stopped:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
