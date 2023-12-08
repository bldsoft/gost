package clickhouse

import (
	"context"
	"errors"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type Batch struct {
	ctx context.Context

	insert string

	conn  driver.Conn
	batch driver.Batch
}

func NewBatch(conn driver.Conn, insertStatement string) (*Batch, error) {
	batch, err := conn.PrepareBatch(context.Background(), insertStatement, driver.WithReleaseConnection())
	if err != nil {
		return nil, err
	}

	return &Batch{
		insert: insertStatement,
		conn:   conn,
		batch:  batch,
		ctx:    context.Background(),
	}, nil
}

func (b *Batch) Append(val any) error {
	err := b.batch.AppendStruct(val)
	if err != nil {
		return nil
	}
	if !errors.Is(err, clickhouse.ErrBatchAlreadySent) {
		return err
	}

	if err := b.reset(); err != nil {
		return err
	}

	return b.batch.Append(val)
}

func (b *Batch) Send() error {
	if err := b.batch.Send(); err != nil {
		return err
	}

	return b.reset()
}

func (b *Batch) reset() error {
	batch, err := b.conn.PrepareBatch(b.ctx, b.insert, driver.WithReleaseConnection())
	if err != nil {
		return err
	}

	b.batch = batch

	return nil
}
