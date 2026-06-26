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
		pfx, err := netip.ParsePrefix(ip)
		if err != nil {
			return netip.Prefix{}, err
		}
		return netip.PrefixFrom(pfx.Addr().Unmap(), pfx.Bits()), nil
	}
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return netip.Prefix{}, err
	}
	return netip.PrefixFrom(addr.Unmap(), addr.Unmap().BitLen()), nil
}

func (s *IPTreeSet) Match(ip netip.Addr) bool {
	return s.items.Lookup(ip.Unmap())
}

func (s *IPTreeSet) Len() int {
	return s.items.Size()
}
