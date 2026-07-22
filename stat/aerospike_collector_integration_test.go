//go:build aerospike_integration

package stat

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"testing"

	gost_aerospike "github.com/bldsoft/gost/cache/v2/aerospike"
	"github.com/stretchr/testify/require"
)

func integrationAerospikeCfg(t *testing.T) gost_aerospike.Config {
	t.Helper()

	host := getenvDefaultLocal("AEROSPIKE_INTEGRATION_HOST", "127.0.0.1")
	portStr := getenvDefaultLocal("AEROSPIKE_INTEGRATION_PORT", "3100")
	ns := getenvDefaultLocal("AEROSPIKE_INTEGRATION_NAMESPACE", "streampool")
	port, err := strconv.Atoi(portStr)
	require.NoError(t, err)

	cfgJSON, err := json.Marshal(map[string]any{
		"Hosts":     []map[string]any{{"Host": host, "Port": port}},
		"Namespace": ns,
		"ConnectionPolicy": map[string]any{
			"ConnectionQueueSize":   128,
			"TimeOutMs":             10_000,
			"IdleTimeoutMs":         1000,
			"MinConnectionsPerNode": 1,
		},
		"WritePolicy": map[string]any{
			"TotalTimeoutMs":        3000,
			"MaxRetries":            2,
			"SleepBetweenRetriesMs": 50,
			"SocketTimeoutMs":       3000,
		},
		"ReadPolicy": map[string]any{
			"TotalTimeoutMs":        3000,
			"MaxRetries":            2,
			"SleepBetweenRetriesMs": 50,
			"SocketTimeoutMs":       3000,
		},
	})
	require.NoError(t, err)

	var cfg gost_aerospike.Config
	require.NoError(t, json.Unmarshal(cfgJSON, &cfg))
	return cfg
}

func getenvDefaultLocal(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func TestAerospikeDistrStatIntegration_Stat(t *testing.T) {
	cfg := integrationAerospikeCfg(t)

	st, err := gost_aerospike.NewStorage(cfg)
	require.NoError(t, err)
	t.Cleanup(func() { st.Close() })

	col := NewAerospikeCollector(st, cfg.Namespace)
	out := col.Stat(context.Background())

	require.Equal(t, "aerospike", out.ServiceType)
	require.Empty(t, out.Err)

	pl, ok := out.Stat.(map[string]any)
	require.True(t, ok)
	require.Contains(t, pl, AerospikeKeyPool)
	require.Contains(t, pl, AerospikeKeyMemUtilization)
	require.Contains(t, pl, AerospikeKeyMemUtilOK)

	pool := pl[AerospikeKeyPool]
	require.NotNil(t, pool)

	_, gotMemOk := pl[AerospikeKeyMemUtilOK].(bool)
	require.True(t, gotMemOk, "mem_utilization_ok should be a bool")

	memU, gotMemU := pl[AerospikeKeyMemUtilization].(float64)
	require.True(t, gotMemU, "mem_utilization should decode as float64: got %T", pl[AerospikeKeyMemUtilization])
	require.GreaterOrEqual(t, memU, 0.0)
	require.LessOrEqual(t, memU, 1.0)
}

func TestAerospikeDistrStatIntegration_MemUtilDisabledWhenNamespaceEmpty(t *testing.T) {
	cfg := integrationAerospikeCfg(t)

	st, err := gost_aerospike.NewStorage(cfg)
	require.NoError(t, err)
	t.Cleanup(func() { st.Close() })

	col := NewAerospikeCollector(st, "")
	out := col.Stat(context.Background())

	require.Empty(t, out.Err)

	pl := out.Stat.(map[string]any)
	memOk, ok := pl[AerospikeKeyMemUtilOK].(bool)
	require.True(t, ok)
	require.False(t, memOk)

	memU := pl[AerospikeKeyMemUtilization].(float64)
	require.InDelta(t, 0, memU, 1e-9)
}
