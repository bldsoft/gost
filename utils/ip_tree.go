package utils

import (
	"net/netip"
	"strings"

	"github.com/gaissmai/bart"
)

type IPTreeSet struct {
	items *bart.Lite
}

func NewIPTreeSet(ipCidrs ...string) *IPTreeSet {
	res := &IPTreeSet{
		items: new(bart.Lite),
	}
	for _, ipCidr := range ipCidrs {
		pfx, err := ipKeyToPrefix(ipCidr)
		if err != nil {
			continue
		}
		res.items.Insert(pfx)
	}

	return res
}

func ipKeyToPrefix(ip string) (netip.Prefix, error) {
	if strings.Contains(ip, "/") {
		return netip.ParsePrefix(ip)
	}
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return netip.Prefix{}, err
	}
	return netip.PrefixFrom(addr, addr.BitLen()), nil
}

func (s *IPTreeSet) Match(ip netip.Addr) bool {
	return s.items.Lookup(ip)
}

func (s *IPTreeSet) Len() int {
	return s.items.Size()
}
