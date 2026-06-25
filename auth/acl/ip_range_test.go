package acl

import (
	"encoding/json"
	"fmt"
	"net"
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
		assert.True(t, acl.ACL.Contains(net.ParseIP("127.0.0.1")))
		assert.True(t, acl.ACL.Contains(net.ParseIP("192.168.0.1")))
		assert.False(t, acl.ACL.Contains(net.ParseIP("192.168.0.2")))
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
	assert.True(t, acl.ACL.Contains(net.ParseIP("127.0.0.1")))
	assert.True(t, acl.ACL.Contains(net.ParseIP("192.168.0.1")))
	assert.False(t, acl.ACL.Contains(net.ParseIP("192.168.0.2")))
}

// Legacy method
func (r IpRange) isInIPs(client net.IP, ips []net.IP) bool {
	for _, ip := range ips {
		if client.Equal(ip) {
			return true
		}
	}
	return false
}

// Legacy method
func (r IpRange) isInSubnets(ip net.IP, subs []*net.IPNet) bool {
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
	target      func(items IpRange, n int) net.IP
}

func benchmarkIPRangeCases() []ipBenchCase {
	mixedScenarios := []struct {
		name   string
		target func(items IpRange, n int) net.IP
	}{
		{"Hit_Early", func(items IpRange, _ int) net.IP {
			return makeIPBenchTarget(items, 0)
		}},
		{"Hit_Middle", func(items IpRange, n int) net.IP {
			return makeIPBenchTarget(items, n/2)
		}},
		{"Hit_Late", func(items IpRange, n int) net.IP {
			return makeIPBenchTarget(items, n-1)
		}},
		{"Miss", func(_ IpRange, _ int) net.IP {
			return net.ParseIP("67.67.67.67")
		}},
	}

	cidrScenarios := []struct {
		name   string
		target func(items IpRange, n int) net.IP
	}{
		{"Hit_First_CIDR", func(_ IpRange, _ int) net.IP {
			return net.ParseIP("10.0.0.50")
		}},
		{"Hit_Middle_CIDR", func(_ IpRange, _ int) net.IP {
			return net.ParseIP("10.68.0.100") // pos=4 → 10.68.0.0/24
		}},
		{"Hit_Last_CIDR", func(_ IpRange, _ int) net.IP {
			return net.ParseIP("2001:db8:27::1") // pos=39
		}},
		{"Miss", func(_ IpRange, _ int) net.IP {
			return net.ParseIP("67.67.67.67")
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

	compareFns := map[string]func(IpRange) func(net.IP) bool{
		"Legacy": func(r IpRange) func(ip net.IP) bool {
			return func(ip net.IP) bool {
				return r.isInSubnets(ip, r.cidrs) || r.isInIPs(ip, r.ips)
			}
		},
		"New": func(r IpRange) func(ip net.IP) bool {
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

func benchmarkIPMatch(b *testing.B, compareFn func(net.IP) bool, target net.IP) {
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

func makeIPBenchTarget(items IpRange, pos int) net.IP {
	var res string
	if pos%3 == 0 {
		octet := (pos * 17) % 256
		res = fmt.Sprintf("10.%d.0.55", octet)
	} else if pos%5 == 0 {
		res = fmt.Sprintf("2001:db8:%x::1", pos)
	} else {
		res = fmt.Sprintf("172.%d.%d.%d", (pos*7)%256, (pos*13)%256, (pos*19)%256)
	}
	return net.ParseIP(res)
}

/**
** BENCHMARK RESULTS **
                                         │  before.txt   │              after.txt              │
                                         │    sec/op     │   sec/op     vs base                │
IPRangeContains/Hit_Early/items_10-12        13.94n ± 2%   20.79n ± 2%  +49.16% (p=0.000 n=10)
IPRangeContains/Hit_Middle/items_10-12       26.20n ± 4%   14.87n ± 8%  -43.24% (p=0.000 n=10)
IPRangeContains/Hit_Late/items_10-12         60.08n ± 2%   21.22n ± 3%  -64.68% (p=0.000 n=10)
IPRangeContains/Miss/items_10-12             73.45n ± 2%   12.30n ± 8%  -83.25% (p=0.000 n=10)
IPRangeContains/CIDR/items_10-12             14.41n ± 3%   22.38n ± 4%  +55.33% (p=0.000 n=10)
IPRangeContains/Hit_Early/items_100-12       14.29n ± 2%   22.48n ± 4%  +57.35% (p=0.000 n=10)
IPRangeContains/Hit_Middle/items_100-12     160.35n ± 3%   33.16n ± 2%  -79.32% (p=0.000 n=10)
IPRangeContains/Hit_Late/items_100-12       535.85n ± 1%   22.04n ± 4%  -95.89% (p=0.000 n=10)
IPRangeContains/Miss/items_100-12           766.55n ± 5%   12.34n ± 2%  -98.39% (p=0.000 n=10)
IPRangeContains/CIDR/items_100-12            14.38n ± 2%   22.17n ± 4%  +54.17% (p=0.000 n=10)
IPRangeContains/Hit_Early/items_1000-12      14.45n ± 2%   21.72n ± 2%  +50.31% (p=0.000 n=10)
IPRangeContains/Hit_Middle/items_1000-12   1349.00n ± 2%   33.53n ± 5%  -97.51% (p=0.000 n=10)
IPRangeContains/Hit_Late/items_1000-12     1225.00n ± 1%   22.00n ± 3%  -98.20% (p=0.000 n=10)
IPRangeContains/Miss/items_1000-12         7392.00n ± 3%   12.34n ± 2%  -99.83% (p=0.000 n=10)
IPRangeContains/CIDR/items_1000-12           14.10n ± 1%   21.94n ± 2%  +55.62% (p=0.000 n=10)
geomean                                      98.37n        20.10n       -79.57%


	Allocs and Byte used are the same
**/
