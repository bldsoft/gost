package clickhouse

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type Batch struct {
	ctx context.Context

	insert string

	conn  driver.Conn
	batch driver.Batch
}

func NewBatch(conn driver.Conn, insertStatement string) (*Batch, error) {
	batch, err := conn.PrepareBatch(context.Background(), insertStatement)
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
	return b.batch.AppendStruct(val)
}

func (b *Batch) Send() error {
	if err := b.batch.Send(); err != nil {
		return err
	}

	return b.reset()
}

func (b *Batch) reset() error {
	batch, err := b.conn.PrepareBatch(b.ctx, b.insert)
	if err != nil {
		return err
	}

	b.batch = batch

	return nil
}
