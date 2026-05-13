package stat

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_namespaceUtilizationFromInfo(t *testing.T) {
	t.Parallel()
	u, ok := namespaceUtilizationFromInfo("")
	require.False(t, ok)

	u, ok = namespaceUtilizationFromInfo("free-pct-memory=40")
	require.True(t, ok)
	require.InEpsilon(t, 0.6, u, 1e-9)

	u, ok = namespaceUtilizationFromInfo("free-pct-disk=25; free-pct-memory=60")
	require.True(t, ok)
	require.InEpsilon(t, 0.75, u, 1e-9) // max(0.75, 0.4)

	u, ok = namespaceUtilizationFromInfo("used-bytes-memory=30; memory-size=100")
	require.True(t, ok)
	require.InEpsilon(t, 0.3, u, 1e-9)

	u, ok = namespaceUtilizationFromInfo("stop-writes=true; free-pct-memory=1")
	require.True(t, ok)
	require.InEpsilon(t, 1, u, 1e-9)
}
