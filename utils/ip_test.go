package utils

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCanonicalAddr(t *testing.T) {
	require.Equal(t, "192.168.0.1", CanonicalAddr(netip.MustParseAddr("::ffff:192.168.0.1")).String())
	require.Equal(t, "192.168.0.1", CanonicalAddr(netip.MustParseAddr("192.168.0.1")).String())
	require.False(t, CanonicalAddr(netip.Addr{}).IsValid())
}

func TestCanonicalPrefix(t *testing.T) {
	pfx, err := CanonicalPrefix(netip.MustParsePrefix("::ffff:10.0.0.0/120"))
	require.NoError(t, err)
	require.Equal(t, "10.0.0.0/24", pfx.String())

	_, err = CanonicalPrefix(netip.MustParsePrefix("::ffff:192.168.0.0/80"))
	require.ErrorIs(t, err, ErrInvalidIP)
}

func TestIPTree_mappedKeysAndLPM(t *testing.T) {
	tree := NewIPTree[string]()
	require.NoError(t, tree.Insert("::ffff:192.168.0.1", "host"))
	require.NoError(t, tree.Insert("10.0.0.0/8", "broad"))
	require.NoError(t, tree.Insert("10.1.0.0/16", "specific"))

	got, ok := tree.Lookup(netip.MustParseAddr("192.168.0.1"))
	require.True(t, ok)
	require.Equal(t, "host", got)

	got, ok = tree.LookupString("::ffff:192.168.0.1")
	require.True(t, ok)
	require.Equal(t, "host", got)

	got, ok = tree.LookupString("10.1.2.3")
	require.True(t, ok)
	require.Equal(t, "specific", got)

	got, ok = tree.LookupString("10.2.0.1")
	require.True(t, ok)
	require.Equal(t, "broad", got)

	require.NoError(t, tree.Delete("::ffff:192.168.0.1"))
	_, ok = tree.LookupString("192.168.0.1")
	require.False(t, ok)
}

func TestIPTreeSet_mappedLookup(t *testing.T) {
	s := NewIPTreeSet("192.168.0.1", "::ffff:10.0.0.0/120")
	require.True(t, s.Match(netip.MustParseAddr("::ffff:192.168.0.1")))
	require.True(t, s.Match(netip.MustParseAddr("10.0.0.55")))
	require.False(t, s.Match(netip.MustParseAddr("11.0.0.1")))
}
