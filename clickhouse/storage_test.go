package clickhouse

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConnectSkipsEmptyDSN(t *testing.T) {
	db := NewStorage(Config{})

	require.NoError(t, db.Connect())
	require.False(t, db.IsReady())
	require.Nil(t, db.nativeConn())
}

func TestConnectReturnsErrorWhenUnreachable(t *testing.T) {
	cfg := Config{Dsn: "clickhouse://127.0.0.1:1/db"}
	require.NoError(t, cfg.Validate())
	cfg.options.DialTimeout = 50 * time.Millisecond

	db := NewStorage(cfg)
	require.Error(t, db.Connect())
	require.False(t, db.IsReady())
	require.Nil(t, db.Db)
	require.Nil(t, db.nativeConn())
}

func TestRunMigrationsWithoutConnection(t *testing.T) {
	db := NewStorage(Config{})

	require.NotPanics(t, db.RunMigrations)
	require.NotPanics(t, db.RunMigrations)
	require.False(t, db.IsReplicationEnabled())
}

func TestNewBatchWithoutConnection(t *testing.T) {
	_, err := NewBatch(nil, "INSERT INTO t VALUES")
	require.ErrorIs(t, err, ErrNotConnected)
}
