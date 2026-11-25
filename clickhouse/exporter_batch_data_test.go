package clickhouse

import (
	"context"
	"errors"
	"testing"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/stretchr/testify/require"
)

var errTableUnavailable = errors.New("table unavailable")

type fakeConn struct {
	driver.Conn

	prepareErr error
	batches    []*fakeBatch
}

func (c *fakeConn) PrepareBatch(context.Context, string, ...driver.PrepareBatchOption) (driver.Batch, error) {
	if c.prepareErr != nil {
		return nil, c.prepareErr
	}
	b := &fakeBatch{}
	c.batches = append(c.batches, b)

	return b, nil
}

type fakeBatch struct {
	driver.Batch

	rows []any
	sent bool
}

func (b *fakeBatch) AppendStruct(v any) error {
	b.rows = append(b.rows, v)

	return nil
}

func (b *fakeBatch) Send() error {
	b.sent = true

	return nil
}

func (b *fakeBatch) IsSent() bool {
	return b.sent
}

type exporterTestRow struct {
	Name string
}

func TestExporterBatchRecoversFromFailedPrepare(t *testing.T) {
	conn := &fakeConn{prepareErr: errTableUnavailable}
	storage := &Storage{}
	storage.setNativeConn(conn)

	var batch *exporterBatch[*exporterTestRow]
	require.NotPanics(t, func() {
		batch = newExporterBatch[*exporterTestRow](storage, "T")
	})
	require.NoError(t, batch.Send())

	_, err := batch.Add(&exporterTestRow{Name: "a"})
	require.ErrorIs(t, err, errTableUnavailable)
	require.Zero(t, batch.Len())

	conn.prepareErr = nil
	n, err := batch.Add(&exporterTestRow{Name: "b"})
	require.NoError(t, err)
	require.Equal(t, 1, n)
	require.NoError(t, batch.Send())

	require.NotEmpty(t, conn.batches)
	require.True(t, conn.batches[0].sent)
	require.Equal(t, []any{&exporterTestRow{Name: "b"}}, conn.batches[0].rows)
}

func TestExporterBatchWaitsForConnection(t *testing.T) {
	storage := &Storage{}

	var batch *exporterBatch[*exporterTestRow]
	require.NotPanics(t, func() {
		batch = newExporterBatch[*exporterTestRow](storage, "T")
	})

	_, err := batch.Add(&exporterTestRow{Name: "a"})
	require.ErrorIs(t, err, ErrNotConnected)
	require.Zero(t, batch.Len())

	conn := &fakeConn{}
	storage.setNativeConn(conn)
	n, err := batch.Add(&exporterTestRow{Name: "b"})
	require.NoError(t, err)
	require.Equal(t, 1, n)
	require.NoError(t, batch.Send())

	require.NotEmpty(t, conn.batches)
	require.True(t, conn.batches[0].sent)
	require.Equal(t, []any{&exporterTestRow{Name: "b"}}, conn.batches[0].rows)
}
