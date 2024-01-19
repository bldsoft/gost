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
	batch, err := conn.PrepareBatch(
		context.Background(),
		insertStatement,
		driver.WithReleaseConnection(),
	)
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

func (b *Batch) Append(val interface{}) error {
	if b.batch.IsSent() {
		if err := b.reset(); err != nil {
			return err
		}
	}
	return b.batch.AppendStruct(val)
}

func (b *Batch) Send() error {
	defer b.reset()

	return b.batch.Send()
}

func (b *Batch) reset() error {
	batch, err := b.conn.PrepareBatch(b.ctx, b.insert, driver.WithReleaseConnection())
	if err != nil {
		return err
	}

	b.batch = batch

	return nil
}
