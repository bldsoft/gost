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
	b := newBatch(conn, insertStatement)
	if err := b.prepareNext(); err != nil {
		return nil, err
	}

	return b, nil
}

func newBatch(conn driver.Conn, insertStatement string) *Batch {
	return &Batch{
		insert: insertStatement,
		conn:   conn,
		ctx:    context.Background(),
	}
}

func (b *Batch) Append(val any) error {
	if b.batch == nil || b.batch.IsSent() {
		if err := b.prepareNext(); err != nil {
			return err
		}
	}

	return b.batch.AppendStruct(val)
}

func (b *Batch) Send() error {
	if b.batch == nil {
		return nil
	}
	defer func() { _ = b.prepareNext() }()

	return b.batch.Send()
}

func (b *Batch) prepareNext() error {
	batch, err := b.conn.PrepareBatch(b.ctx, b.insert, driver.WithReleaseConnection())
	if err != nil {
		return err
	}

	b.batch = batch

	return nil
}
