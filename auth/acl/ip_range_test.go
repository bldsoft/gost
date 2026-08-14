package acl

import (
	"encoding/json"
	"fmt"
	"net/netip"
	"slices"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

type acl struct {
	ACL IpRange `json:"acl" bson:"acl"`
}

func TestIpRangeJson(t *testing.T) {
	var got acl
	t.Run("Unmarshal", func(t *testing.T) {
		assert.NoError(t, json.Unmarshal([]byte(`{ "acl": ["127.0.0.0/24","192.168.0.1"]}`), &got))
		assert.True(t, got.ACL.Contains(netip.MustParseAddr("127.0.0.1")))
		assert.True(t, got.ACL.Contains(netip.MustParseAddr("192.168.0.1")))
		assert.False(t, got.ACL.Contains(netip.MustParseAddr("192.168.0.2")))
	})
	t.Run("Marshal", func(t *testing.T) {
		data, err := json.Marshal(got)
		assert.NoError(t, err)
		assert.Equal(t, `{"acl":["192.168.0.1","127.0.0.0/24"]}`, string(data))
	})
	t.Run("Unmasked CIDR roundtrip", func(t *testing.T) {
		var masked acl
		assert.NoError(t, json.Unmarshal([]byte(`{"acl":["10.0.0.1/24"]}`), &masked))
		assert.Equal(t, []string{"10.0.0.0/24"}, masked.ACL.Strings())
		data, err := json.Marshal(masked)
		assert.NoError(t, err)
		assert.Equal(t, `{"acl":["10.0.0.0/24"]}`, string(data))
	})
	t.Run("Mapped address roundtrip", func(t *testing.T) {
		var mapped acl
		assert.NoError(t, json.Unmarshal([]byte(`{"acl":["::ffff:192.168.0.1","::ffff:10.0.0.0/120"]}`), &mapped))
		assert.Equal(t, []string{"192.168.0.1", "10.0.0.0/24"}, mapped.ACL.Strings())
	})
}

func TestIpRangeBson(t *testing.T) {
	acl := acl{
		ACL: MustIpRangeFromStrings("127.0.0.0/24", "192.168.0.1"),
	}

	data, err := bson.Marshal(acl)
	assert.NoError(t, err)

	assert.NoError(t, bson.Unmarshal(data, &acl))
	assert.True(t, acl.ACL.Contains(netip.MustParseAddr("127.0.0.1")))
	assert.True(t, acl.ACL.Contains(netip.MustParseAddr("192.168.0.1")))
	assert.False(t, acl.ACL.Contains(netip.MustParseAddr("192.168.0.2")))
}

func TestIpRangeIPv4MappedIPv6(t *testing.T) {
	t.Run("IPv4 stored, IPv4-mapped IPv6 lookup", func(t *testing.T) {
		ipRange := MustIpRangeFromStrings("192.168.0.1", "10.0.0.0/24")
		assert.True(t, ipRange.Contains(netip.MustParseAddr("::ffff:192.168.0.1")))
		assert.True(t, ipRange.Contains(netip.MustParseAddr("::ffff:10.0.0.50")))
		assert.False(t, ipRange.Contains(netip.MustParseAddr("::ffff:192.168.0.2")))
	})

	t.Run("IPv4-mapped IPv6 stored, IPv4 lookup", func(t *testing.T) {
		ipRange := MustIpRangeFromStrings("::ffff:192.168.0.1")
		assert.True(t, ipRange.Contains(netip.MustParseAddr("192.168.0.1")))
		assert.False(t, ipRange.Contains(netip.MustParseAddr("192.168.0.2")))
	})

	t.Run("Mixed storage and lookup", func(t *testing.T) {
		ipRange := MustIpRangeFromStrings("192.168.1.1", "::ffff:192.168.2.1", "10.0.0.0/24")
		assert.True(t, ipRange.Contains(netip.MustParseAddr("192.168.1.1")))
		assert.True(t, ipRange.Contains(netip.MustParseAddr("::ffff:192.168.1.1")))
		assert.True(t, ipRange.Contains(netip.MustParseAddr("192.168.2.1")))
		assert.True(t, ipRange.Contains(netip.MustParseAddr("::ffff:192.168.2.1")))
		assert.True(t, ipRange.Contains(netip.MustParseAddr("10.0.0.100")))
		assert.True(t, ipRange.Contains(netip.MustParseAddr("::ffff:10.0.0.100")))
	})
}

func TestIpRangeEmptyContains(t *testing.T) {
	var zero IpRange
	assert.True(t, zero.Empty())
	assert.False(t, zero.Contains(netip.MustParseAddr("127.0.0.1")))

	empty := MustIpRangeFromStrings()
	assert.True(t, empty.Empty())
	assert.False(t, empty.Contains(netip.MustParseAddr("127.0.0.1")))
}

func TestIpRangeUnmaskedCIDR(t *testing.T) {
	ipRange := MustIpRangeFromStrings("10.0.0.1/24")
	assert.Equal(t, []string{"10.0.0.0/24"}, ipRange.Strings())
	assert.True(t, ipRange.Contains(netip.MustParseAddr("10.0.0.50")))
	assert.False(t, ipRange.Contains(netip.MustParseAddr("10.0.1.1")))
}

func TestIpRangeIPv6(t *testing.T) {
	ipRange := MustIpRangeFromStrings("2001:db8::1", "2001:db8:1::/48")
	assert.True(t, ipRange.Contains(netip.MustParseAddr("2001:db8::1")))
	assert.True(t, ipRange.Contains(netip.MustParseAddr("2001:db8:1::ffff")))
	assert.False(t, ipRange.Contains(netip.MustParseAddr("2001:db8:2::1")))
}

func TestIpRangeZonedIPv6(t *testing.T) {
	ipRange := MustIpRangeFromStrings("fe80::1%eth0")
	assert.True(t, ipRange.Contains(netip.MustParseAddr("fe80::1")))
	assert.True(t, ipRange.Contains(netip.MustParseAddr("fe80::1%eth0")))
	assert.False(t, ipRange.Contains(netip.MustParseAddr("fe80::2")))
}

func TestIpRangeParseErrors(t *testing.T) {
	_, err := IpRangeFromStrings("not-an-ip")
	assert.Error(t, err)
	_, err = IpRangeFromStrings("10.0.0.0/99")
	assert.Error(t, err)
}

func TestIpRangeIPsAndCIDRsCopy(t *testing.T) {
	ipRange := MustIpRangeFromStrings("192.168.0.1", "10.0.0.0/24")
	ips := ipRange.IPs()
	cidrs := ipRange.CIDRs()
	ips[0] = netip.MustParseAddr("1.1.1.1")
	cidrs[0] = netip.MustParsePrefix("8.8.8.0/24")
	assert.True(t, ipRange.Contains(netip.MustParseAddr("192.168.0.1")))
	assert.True(t, ipRange.Contains(netip.MustParseAddr("10.0.0.50")))
	assert.False(t, ipRange.Contains(netip.MustParseAddr("1.1.1.1")))
}

func TestIpRangeTreeMatchesLinear(t *testing.T) {
	fixed := []netip.Addr{
		netip.MustParseAddr("10.0.0.50"),
		netip.MustParseAddr("10.68.0.100"),
		netip.MustParseAddr("67.67.67.67"),
		netip.MustParseAddr("2001:db8:27::1"),
		netip.MustParseAddr("172.16.0.1"),
	}
	for _, n := range []int{10, 50, 100} {
		for _, gen := range []func(int) IpRange{makeIPRangeBench, makeIPRangeBenchCIDROnly} {
			r := gen(n)
			probes := append([]netip.Addr{}, fixed...)
			for i := range n {
				probes = append(probes, makeIPBenchTarget(r, i))
			}
			for _, ip := range probes {
				legacy := r.isInSubnets(ip, r.cidrs) || r.isInIPs(ip, r.ips)
				assert.Equalf(t, legacy, r.Contains(ip), "n=%d ip=%s", n, ip)
			}
		}
	}
}

// Legacy method
func (r IpRange) isInIPs(client netip.Addr, ips []netip.Addr) bool {
	return slices.Contains(ips, client)
}

// Legacy method
func (r IpRange) isInSubnets(ip netip.Addr, subs []netip.Prefix) bool {
	for _, subnet := range subs {
		if subnet.Contains(ip) {
			return true
		}
	}
	return false
}

type ipBenchCase struct {
	name        string
	generatorFn func(n int) IpRange
	n           int
	target      func(items IpRange, n int) netip.Addr
}

func benchmarkIPRangeCases() []ipBenchCase {
	mixedScenarios := []struct {
		name   string
		target func(items IpRange, n int) netip.Addr
	}{
		{"Hit_Early", func(items IpRange, _ int) netip.Addr {
			return makeIPBenchTarget(items, 0)
		}},
		{"Hit_Middle", func(items IpRange, n int) netip.Addr {
			return makeIPBenchTarget(items, n/2)
		}},
		{"Hit_Late", func(items IpRange, n int) netip.Addr {
			return makeIPBenchTarget(items, n-1)
		}},
		{"Miss", func(_ IpRange, _ int) netip.Addr {
			return netip.MustParseAddr("67.67.67.67")
		}},
	}

	cidrScenarios := []struct {
		name   string
		target func(items IpRange, n int) netip.Addr
	}{
		{"Hit_First_CIDR", func(_ IpRange, _ int) netip.Addr {
			return netip.MustParseAddr("10.0.0.50")
		}},
		{"Hit_Middle_CIDR", func(_ IpRange, _ int) netip.Addr {
			return netip.MustParseAddr("10.68.0.100") // pos=4 → 10.68.0.0/24
		}},
		{"Hit_Last_CIDR", func(_ IpRange, _ int) netip.Addr {
			return netip.MustParseAddr("2001:db8:27::1") // pos=39
		}},
		{"Miss", func(_ IpRange, _ int) netip.Addr {
			return netip.MustParseAddr("67.67.67.67")
		}},
	}

	var cases []ipBenchCase
	for _, n := range []int{10, 50, 100} {
		for _, sc := range mixedScenarios {
			cases = append(cases, ipBenchCase{sc.name, makeIPRangeBench, n, sc.target})
		}
		for _, sc := range cidrScenarios {
			cases = append(cases, ipBenchCase{sc.name, makeIPRangeBenchCIDROnly, n, sc.target})
		}
	}
	return cases
}

func BenchmarkIPRangeContains(b *testing.B) {
	cases := benchmarkIPRangeCases()

	compareFns := map[string]func(IpRange) func(netip.Addr) bool{
		"Legacy": func(r IpRange) func(ip netip.Addr) bool {
			return func(ip netip.Addr) bool {
				return r.isInSubnets(ip, r.cidrs) || r.isInIPs(ip, r.ips)
			}
		},
		"New": func(r IpRange) func(ip netip.Addr) bool {
			return r.Contains
		},
	}

	for _, implName := range []string{"Legacy", "New"} {
		b.Run(implName, func(b *testing.B) {
			for _, c := range cases {
				b.Run(c.name+"/items_"+strconv.Itoa(c.n), func(b *testing.B) {
					items := c.generatorFn(c.n)
					benchmarkIPMatch(b, compareFns[implName](items), c.target(items, c.n))
				})
			}
		})
	}
}

func benchmarkIPMatch(b *testing.B, compareFn func(netip.Addr) bool, target netip.Addr) {
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		compareFn(target)
	}
}

func makeIPRangeBench(n int) IpRange {
	items := make([]string, 0, n)
	for i := range n {
		var key string
		if i%3 == 0 {
			key = fmt.Sprintf("10.%d.0.0/24", (i*17)%256)
		} else if i%5 == 0 {
			key = fmt.Sprintf("2001:db8:%x::/48", i)
		} else {
			key = fmt.Sprintf("172.%d.%d.%d", (i*7)%256, (i*13)%256, (i*19)%256)
		}
		items = append(items, key)
	}
	return MustIpRangeFromStrings(items...)
}

func makeIPRangeBenchCIDROnly(n int) IpRange {
	items := make([]string, 0, n)
	for i := range n {
		var key string
		if i%2 == 0 {
			key = fmt.Sprintf("10.%d.0.0/24", (i*17)%256)
		} else {
			key = fmt.Sprintf("2001:db8:%x::/48", i)
		}
		items = append(items, key)
	}
	return MustIpRangeFromStrings(items...)
}

func makeIPBenchTarget(items IpRange, pos int) netip.Addr {
	var res string
	if pos%3 == 0 {
		octet := (pos * 17) % 256
		res = fmt.Sprintf("10.%d.0.55", octet)
	} else if pos%5 == 0 {
		res = fmt.Sprintf("2001:db8:%x::1", pos)
	} else {
		res = fmt.Sprintf("172.%d.%d.%d", (pos*7)%256, (pos*13)%256, (pos*19)%256)
	}
	ip, _ := netip.ParseAddr(res)
	return ip
}
