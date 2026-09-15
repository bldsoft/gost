package stat

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_namespaceUtilizationFromInfo(t *testing.T) {
	t.Parallel()
	u, ok := namespaceUtilizationFromInfo("")
	require.False(t, ok)

	u, ok = namespaceUtilizationFromInfo("memory_free_pct=40")
	require.True(t, ok)
	require.InEpsilon(t, 0.6, u, 1e-9)

	u, ok = namespaceUtilizationFromInfo("device_free_pct=25; memory_free_pct=60")
	require.True(t, ok)
	require.InEpsilon(t, 0.75, u, 1e-9) // max(0.75, 0.4)

	u, ok = namespaceUtilizationFromInfo("device_used_bytes=30; device_total_bytes=100")
	require.True(t, ok)
	require.InEpsilon(t, 0.3, u, 1e-9)

	u, ok = namespaceUtilizationFromInfo("stop_writes=true; memory_free_pct=1")
	require.True(t, ok)
	require.InEpsilon(t, 1, u, 1e-9)
}

func Test_namespaceUtilizationFromInfo_MetricNames(t *testing.T) {
	t.Parallel()

	t.Run("stop_writes metric", func(t *testing.T) {
		u, ok := namespaceUtilizationFromInfo("stop_writes=true")
		require.True(t, ok)
		require.Equal(t, 1.0, u)
	})

	t.Run("memory_free_pct metric", func(t *testing.T) {
		u, ok := namespaceUtilizationFromInfo("memory_free_pct=30")
		require.True(t, ok)
		require.InEpsilon(t, 0.7, u, 1e-9)
	})

	t.Run("device_free_pct metric", func(t *testing.T) {
		u, ok := namespaceUtilizationFromInfo("device_free_pct=20")
		require.True(t, ok)
		require.InEpsilon(t, 0.8, u, 1e-9)
	})

	t.Run("device_used_bytes and device_total_bytes metrics", func(t *testing.T) {
		u, ok := namespaceUtilizationFromInfo("device_used_bytes=5000000000; device_total_bytes=10000000000")
		require.True(t, ok)
		require.InEpsilon(t, 0.5, u, 1e-9)
	})

	t.Run("combined metrics - takes maximum", func(t *testing.T) {
		u, ok := namespaceUtilizationFromInfo("memory_free_pct=50; device_free_pct=10")
		require.True(t, ok)
		require.InEpsilon(t, 0.9, u, 1e-9) // max(0.5, 0.9) = 0.9
	})

	t.Run("stop_writes overrides everything", func(t *testing.T) {
		u, ok := namespaceUtilizationFromInfo("stop_writes=true; memory_free_pct=90; device_free_pct=80")
		require.True(t, ok)
		require.Equal(t, 1.0, u)
	})
}
