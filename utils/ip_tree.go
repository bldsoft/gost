package utils

import (
	"net"
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

func (s *IPTreeSet) Match(ip net.IP) bool {
	if len(ip) == 4 {
		addr := netip.AddrFrom4([4]byte(ip))
		return s.items.Lookup(addr)
	}

	var (
		addr netip.Addr
		ok   bool
	)

	if v4 := ip.To4(); v4 != nil {
		addr, ok = netip.AddrFromSlice(v4)
	} else {
		v6 := ip.To16()
		addr, ok = netip.AddrFromSlice(v6)
	}
	if !ok {
		return false
	}

	return s.items.Lookup(addr)
}

func (s *IPTreeSet) Len() int {
	return s.items.Size()
}
