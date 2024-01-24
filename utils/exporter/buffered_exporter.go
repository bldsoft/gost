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

type (
	Fields = map[string]interface{}
	Logger interface {
		TraceOrErrorfWithFields(err error, fields Fields, format string, v ...interface{})
	}
)

type nullLogger struct{}

func (nullLogger) TraceOrErrorfWithFields(err error, fields Fields, format string, v ...interface{}) {
	// do nothing
}

type BufferedExporterConfig struct {
	MaxFlushInterval time.Duration
	MaxBatchSize     int
	ChanBufSize      int  // buffer between writer and the goroutine that actually exports data
	PreserveOld      bool // false - discard old data in case of overflow, true - discard new

	Logger Logger
}

type BufferedExporter[T any] struct {
	cfg BufferedExporterConfig

	exportedData Data[T]
	writeC       chan T
	ringBuf      *ringbuf.RingBuf[T]

	stop    chan struct{}
	stopped chan struct{}
}

func NewBuffered[T any](
	data Data[T],
	cfg BufferedExporterConfig,
) *BufferedExporter[T] {
	if cfg.ChanBufSize <= 0 {
		cfg.ChanBufSize = DefaultChanBufSize
	}
	if cfg.Logger == nil {
		cfg.Logger = nullLogger{}
	}
	return &BufferedExporter[T]{
		exportedData: data,
		cfg:          cfg,
		writeC:       make(chan T, cfg.ChanBufSize),
		stop:         make(chan struct{}),
		stopped:      make(chan struct{}),
	}
}

func NewBufferedFromExporter[T any](exporter Exporter[T], cfg BufferedExporterConfig) *BufferedExporter[T] {
	return NewBuffered(NewSlice[T](exporter), cfg)
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

func (be *BufferedExporter[T]) flush() (n int, err error) {
	if be.ringBuf.Empty() && be.exportedData.Len() == 0 {
		return 0, nil
	}
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

	if err := be.fillExportedData(); err != nil {
		return 0, err
	}

	exported := be.exportedData.Len()
	err = be.exportedData.Send()
	if err != nil {
		return 0, err
	}
	return exported, be.exportedData.Reset()
}

func (be *BufferedExporter[T]) fillExportedData() error {
	pullN := min(be.ringBuf.Len(), be.MaxBatchSize()-be.exportedData.Len())
	for i := 0; i < pullN; i++ {
		item, _ := be.ringBuf.Top()
		if _, err := be.exportedData.Add(item); err != nil {
			return err
		}
		be.ringBuf.Remove(1)
	}
	return nil
}

func (be *BufferedExporter[T]) Run() error {
	defer close(be.stopped)

	be.ringBuf = ringbuf.New[T](be.cfg.ChanBufSize).WithOverwrite(!be.cfg.PreserveOld)

	ticker := time.NewTicker(be.MaxFlushInterval())
	defer ticker.Stop()

	flush := func() error {
		n, err := be.flush()
		be.cfg.Logger.TraceOrErrorfWithFields(err, Fields{
			"queued":         len(be.writeC),
			"ring buf":       be.ringBuf.Len(),
			"exported batch": n,
		}, "buffered exporter")
		return err
	}

	flushAll := func() error {
		for !be.ringBuf.Empty() || be.exportedData.Len() > 0 {
			if err := flush(); err != nil {
				return err
			}
		}
		return nil
	}

	lastFlushTriggeredBySize := time.Now()
	for {
		select {
		case item := <-be.writeC:
			_ = be.ringBuf.Push(item)

			if be.ringBuf.Len() == be.MaxBatchSize() {
				_ = flush()
				lastFlushTriggeredBySize = time.Now()
			}
		case <-ticker.C:
			if time.Since(lastFlushTriggeredBySize) >= be.MaxFlushInterval() {
				_ = flushAll()
			}
		case <-be.stop:
			close(be.writeC)

			for item := range be.writeC {
				if !be.ringBuf.Full() {
					_ = be.ringBuf.Push(item)
					continue
				}
				if err := flushAll(); err != nil {
					return err
				}
			}
			return flushAll()
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
