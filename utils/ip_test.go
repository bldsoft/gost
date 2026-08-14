package utils

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCanonicalAddr(t *testing.T) {
	require.Equal(t, "192.168.0.1", CanonicalAddr(netip.MustParseAddr("::ffff:192.168.0.1")).String())
	require.Equal(t, "192.168.0.1", CanonicalAddr(netip.MustParseAddr("192.168.0.1")).String())
	require.Equal(t, "fe80::1", CanonicalAddr(netip.MustParseAddr("fe80::1%eth0")).String())
	require.False(t, CanonicalAddr(netip.Addr{}).IsValid())
}

func TestCanonicalPrefix(t *testing.T) {
	t.Run("mapped IPv6", func(t *testing.T) {
		pfx, err := CanonicalPrefix(netip.MustParsePrefix("::ffff:10.0.0.0/120"))
		require.NoError(t, err)
		require.Equal(t, "10.0.0.0/24", pfx.String())
	})

	t.Run("too-short mapped prefix", func(t *testing.T) {
		_, err := CanonicalPrefix(netip.MustParsePrefix("::ffff:192.168.0.0/80"))
		require.ErrorIs(t, err, ErrInvalidIP)
	})

	t.Run("masks host bits", func(t *testing.T) {
		pfx, err := CanonicalPrefix(netip.MustParsePrefix("10.0.0.1/24"))
		require.NoError(t, err)
		require.Equal(t, "10.0.0.0/24", pfx.String())
	})

	t.Run("mapped /96 is all IPv4", func(t *testing.T) {
		pfx, err := CanonicalPrefix(netip.MustParsePrefix("::ffff:10.1.2.3/96"))
		require.NoError(t, err)
		require.Equal(t, "0.0.0.0/0", pfx.String())
	})

	t.Run("invalid prefix", func(t *testing.T) {
		_, err := CanonicalPrefix(netip.Prefix{})
		require.ErrorIs(t, err, ErrInvalidIP)
	})
}

func TestIPKeyToPrefix(t *testing.T) {
	t.Run("host", func(t *testing.T) {
		pfx, err := IPKeyToPrefix("192.168.0.1")
		require.NoError(t, err)
		require.Equal(t, "192.168.0.1/32", pfx.String())
	})

	t.Run("zoned IPv6 host", func(t *testing.T) {
		pfx, err := IPKeyToPrefix("fe80::1%eth0")
		require.NoError(t, err)
		require.Equal(t, "fe80::1/128", pfx.String())
	})

	t.Run("unmasked CIDR", func(t *testing.T) {
		pfx, err := IPKeyToPrefix("10.0.0.1/24")
		require.NoError(t, err)
		require.Equal(t, "10.0.0.0/24", pfx.String())
	})

	t.Run("invalid", func(t *testing.T) {
		_, err := IPKeyToPrefix("not-an-ip")
		require.Error(t, err)
	})
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

func TestIPTree_PutIPsAndPrefixes(t *testing.T) {
	tree := NewIPTree[string]()
	require.NoError(t, tree.PutIPs("host",
		netip.MustParseAddr("::ffff:192.168.0.1"),
		netip.MustParseAddr("fe80::1%eth0"),
	))
	require.NoError(t, tree.PutPrefixes("net",
		netip.MustParsePrefix("10.0.0.1/24"),
	))

	got, ok := tree.Lookup(netip.MustParseAddr("192.168.0.1"))
	require.True(t, ok)
	require.Equal(t, "host", got)

	got, ok = tree.Lookup(netip.MustParseAddr("fe80::1"))
	require.True(t, ok)
	require.Equal(t, "host", got)

	got, ok = tree.Lookup(netip.MustParseAddr("10.0.0.50"))
	require.True(t, ok)
	require.Equal(t, "net", got)

	require.Error(t, tree.PutIPs("bad", netip.Addr{}))
	require.Error(t, tree.PutPrefixes("bad", netip.Prefix{}))
	require.Error(t, tree.Insert("not-an-ip", "x"))
}

func TestIPTreeSet_mappedLookup(t *testing.T) {
	s := NewIPTreeSet()
	require.NoError(t, s.Put("192.168.0.1", "::ffff:10.0.0.0/120"))
	require.True(t, s.Match(netip.MustParseAddr("::ffff:192.168.0.1")))
	require.True(t, s.Match(netip.MustParseAddr("10.0.0.55")))
	require.False(t, s.Match(netip.MustParseAddr("11.0.0.1")))
}

func TestIPTreeSet_PutIPsAndPrefixes(t *testing.T) {
	s := NewIPTreeSet()
	require.NoError(t, s.PutIPs(
		netip.MustParseAddr("192.168.0.1"),
		netip.MustParseAddr("fe80::1%eth0"),
	))
	require.NoError(t, s.PutPrefixes(netip.MustParsePrefix("10.1.2.3/16")))

	require.True(t, s.Match(netip.MustParseAddr("192.168.0.1")))
	require.True(t, s.Match(netip.MustParseAddr("fe80::1")))
	require.True(t, s.Match(netip.MustParseAddr("10.1.9.9")))
	require.False(t, s.Match(netip.MustParseAddr("10.2.0.1")))
	require.Equal(t, 3, s.Len())

	require.Error(t, s.Put("not-an-ip"))
	require.Error(t, s.PutIPs(netip.Addr{}))
	require.Error(t, s.PutPrefixes(netip.Prefix{}))
}

func TestIPTreeSet_nilSafe(t *testing.T) {
	var s *IPTreeSet
	require.False(t, s.Match(netip.MustParseAddr("1.1.1.1")))
	require.Equal(t, 0, s.Len())
}
