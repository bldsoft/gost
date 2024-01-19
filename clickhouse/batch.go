package clickhouse

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type Batch[T any] struct {
	ctx context.Context

	insert string

	conn  driver.Conn
	batch driver.Batch
}

func NewBatch[T any](conn driver.Conn, insertStatement string) (*Batch[T], error) {
	batch, err := conn.PrepareBatch(context.Background(), insertStatement, driver.WithReleaseConnection())
	if err != nil {
		return nil, err
	}

	return &Batch[T]{
		insert: insertStatement,
		conn:   conn,
		batch:  batch,
		ctx:    context.Background(),
	}, nil
}

func (b *Batch[T]) Append(val T) error {
	if b.batch.IsSent() {
		if err := b.reset(); err != nil {
			return err
		}
	}
	return b.batch.AppendStruct(val)
}

func (b *Batch[T]) Send() error {
	defer b.reset()

	return b.batch.Send()
}

func (b *Batch[T]) reset() error {
	batch, err := b.conn.PrepareBatch(b.ctx, b.insert, driver.WithReleaseConnection())
	if err != nil {
		return err
	}

	b.batch = batch

	return nil
}
