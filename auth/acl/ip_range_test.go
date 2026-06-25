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
	var acl acl
	t.Run("Unmarshal", func(t *testing.T) {
		assert.NoError(t, json.Unmarshal([]byte(`{ "acl": ["127.0.0.0/24","192.168.0.1"]}`), &acl))
		assert.True(t, acl.ACL.Contains(netip.MustParseAddr("127.0.0.1")))
		assert.True(t, acl.ACL.Contains(netip.MustParseAddr("192.168.0.1")))
		assert.False(t, acl.ACL.Contains(netip.MustParseAddr("192.168.0.2")))
	})
	t.Run("Marshal", func(t *testing.T) {
		data, err := json.Marshal(acl)
		assert.NoError(t, err)
		assert.Equal(t, `{"acl":["192.168.0.1","127.0.0.0/24"]}`, string(data))
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

/**
** BENCHMARK RESULTS **
** New/net.IP vs New/netip.Addr **
                                             │   after.txt   │              after2.txt              │
                                             │    sec/op     │    sec/op     vs base                │
IPRangeContains/Hit_Early/items_10-12           21.69n ±  3%   14.41n ±  3%  -33.59% (p=0.000 n=10)
IPRangeContains/Hit_Middle/items_10-12          15.16n ±  3%   13.86n ± 11%   -8.61% (p=0.000 n=10)
IPRangeContains/Hit_Late/items_10-12            21.33n ±  3%   14.89n ±  3%  -30.17% (p=0.000 n=10)
IPRangeContains/Miss/items_10-12               12.165n ±  2%   4.828n ±  2%  -60.31% (p=0.000 n=10)
IPRangeContains/Hit_First_CIDR/items_10-12      21.41n ±  1%   14.71n ±  4%  -31.25% (p=0.000 n=10)
IPRangeContains/Hit_Middle_CIDR/items_10-12     21.60n ±  3%   14.54n ±  3%  -32.67% (p=0.000 n=10)
IPRangeContains/Hit_Last_CIDR/items_10-12       29.02n ± 35%   25.96n ±  2%  -10.54% (p=0.000 n=10)
IPRangeContains/Miss/items_10#01-12            16.650n ± 18%   4.878n ±  3%  -70.70% (p=0.000 n=10)
IPRangeContains/Hit_Early/items_50-12           22.69n ± 10%   15.85n ± 15%  -30.15% (p=0.000 n=10)
IPRangeContains/Hit_Middle/items_50-12          32.49n ±  2%   32.67n ±  7%        ~ (p=0.853 n=10)
IPRangeContains/Hit_Late/items_50-12            3.816n ±  4%   4.437n ±  4%  +16.29% (p=0.000 n=10)
IPRangeContains/Miss/items_50-12               12.135n ±  6%   4.899n ±  3%  -59.63% (p=0.000 n=10)
IPRangeContains/Hit_First_CIDR/items_50-12      22.60n ±  4%   14.52n ±  2%  -35.75% (p=0.000 n=10)
IPRangeContains/Hit_Middle_CIDR/items_50-12     21.93n ±  3%   14.51n ±  3%  -33.82% (p=0.000 n=10)
IPRangeContains/Hit_Last_CIDR/items_50-12       32.50n ±  4%   30.40n ±  2%   -6.46% (p=0.000 n=10)
IPRangeContains/Miss/items_50#01-12            11.945n ±  8%   4.821n ±  7%  -59.64% (p=0.000 n=10)
IPRangeContains/Hit_Early/items_100-12          21.25n ±  3%   15.77n ±  7%  -25.79% (p=0.000 n=10)
IPRangeContains/Hit_Middle/items_100-12         32.21n ±  4%   31.11n ±  2%   -3.40% (p=0.001 n=10)
IPRangeContains/Hit_Late/items_100-12           21.50n ±  4%   14.74n ±  2%  -31.44% (p=0.000 n=10)
IPRangeContains/Miss/items_100-12              12.145n ±  3%   4.917n ±  3%  -59.52% (p=0.000 n=10)
IPRangeContains/Hit_First_CIDR/items_100-12     22.04n ±  3%   14.64n ±  4%  -33.56% (p=0.000 n=10)
IPRangeContains/Hit_Middle_CIDR/items_100-12    21.29n ±  2%   15.00n ±  5%  -29.55% (p=0.000 n=10)
IPRangeContains/Hit_Last_CIDR/items_100-12      31.58n ±  2%   29.99n ±  5%   -5.00% (p=0.001 n=10)
IPRangeContains/Miss/items_100#01-12           11.790n ±  3%   4.840n ±  3%  -58.95% (p=0.000 n=10)
geomean                                         18.82n         12.33n        -34.48%


	Allocs and Byte used are the same


** Legacy/netip.Addr vs New/netip.Addr **
                                             │  before2.txt   │               after2.txt               │
                                             │     sec/op     │    sec/op      vs base                 │
IPRangeContains/Hit_Early/items_10-12            5.390n ±  7%   14.405n ±  3%  +167.23% (p=0.000 n=10)
IPRangeContains/Hit_Middle/items_10-12           12.54n ±  5%    13.86n ± 11%   +10.48% (p=0.027 n=10)
IPRangeContains/Hit_Late/items_10-12             22.07n ±  3%    14.89n ±  3%   -32.53% (p=0.000 n=10)
IPRangeContains/Miss/items_10-12                29.185n ± 39%    4.828n ±  2%   -83.46% (p=0.000 n=10)
IPRangeContains/Hit_First_CIDR/items_10-12       6.516n ± 14%   14.715n ±  4%  +125.85% (p=0.000 n=10)
IPRangeContains/Hit_Middle_CIDR/items_10-12      24.46n ± 34%    14.54n ±  3%   -40.56% (p=0.000 n=10)
IPRangeContains/Hit_Last_CIDR/items_10-12        42.80n ± 16%    25.96n ±  2%   -39.35% (p=0.000 n=10)
IPRangeContains/Miss/items_10#01-12             40.485n ±  2%    4.878n ±  3%   -87.95% (p=0.000 n=10)
IPRangeContains/Hit_Early/items_50-12            5.164n ±  6%   15.850n ± 15%  +206.93% (p=0.000 n=10)
IPRangeContains/Hit_Middle/items_50-12           48.33n ±  1%    32.67n ±  7%   -32.41% (p=0.000 n=10)
IPRangeContains/Hit_Late/items_50-12            89.450n ±  2%    4.437n ±  4%   -95.04% (p=0.000 n=10)
IPRangeContains/Miss/items_50-12               112.200n ±  1%    4.899n ±  3%   -95.63% (p=0.000 n=10)
IPRangeContains/Hit_First_CIDR/items_50-12       5.151n ±  1%   14.520n ±  2%  +181.89% (p=0.000 n=10)
IPRangeContains/Hit_Middle_CIDR/items_50-12      21.21n ±  1%    14.51n ±  3%   -31.59% (p=0.000 n=10)
IPRangeContains/Hit_Last_CIDR/items_50-12       168.85n ±  2%    30.40n ±  2%   -81.99% (p=0.000 n=10)
IPRangeContains/Miss/items_50#01-12            197.500n ±  3%    4.821n ±  7%   -97.56% (p=0.000 n=10)
IPRangeContains/Hit_Early/items_100-12           5.146n ±  2%   15.770n ±  7%  +206.45% (p=0.000 n=10)
IPRangeContains/Hit_Middle/items_100-12          87.73n ±  4%    31.11n ±  2%   -64.54% (p=0.000 n=10)
IPRangeContains/Hit_Late/items_100-12           192.25n ±  1%    14.74n ±  2%   -92.33% (p=0.000 n=10)
IPRangeContains/Miss/items_100-12              222.700n ±  2%    4.917n ±  3%   -97.79% (p=0.000 n=10)
IPRangeContains/Hit_First_CIDR/items_100-12      5.178n ±  1%   14.640n ±  4%  +182.76% (p=0.000 n=10)
IPRangeContains/Hit_Middle_CIDR/items_100-12     21.32n ±  2%    15.00n ±  5%   -29.67% (p=0.000 n=10)
IPRangeContains/Hit_Last_CIDR/items_100-12      168.15n ±  2%    29.99n ±  5%   -82.16% (p=0.000 n=10)
IPRangeContains/Miss/items_100#01-12           383.800n ±  1%    4.840n ±  3%   -98.74% (p=0.000 n=10)
geomean                                          35.41n          12.33n         -65.17%
**/
