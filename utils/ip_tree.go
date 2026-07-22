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
	s := &IPTreeSet{
		items: new(bart.Lite),
	}
	for _, ipCidr := range ipCidrs {
		pfx, err := s.ipKeyToPrefix(ipCidr)
		if err != nil {
			continue
		}
		s.items.Insert(pfx)
	}

	return s
}

func (s *IPTreeSet) ipKeyToPrefix(ip string) (netip.Prefix, error) {
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

func (s *IPTreeSet) Put(ipCidrs ...string) error {
	for _, ipCidr := range ipCidrs {
		pfx, err := s.ipKeyToPrefix(ipCidr)
		if err != nil {
			return err
		}
		s.items.Insert(pfx)
	}
	return nil
}

func (s *IPTreeSet) Delete(ipCidrs ...string) error {
	for _, ipCidr := range ipCidrs {
		pfx, err := s.ipKeyToPrefix(ipCidr)
		if err != nil {
			return err
		}
		s.items.Delete(pfx)
	}
	return nil
}

func (s *IPTreeSet) Match(ip netip.Addr) bool {
	return s.items.Lookup(ip.Unmap())
}

func (s *IPTreeSet) Len() int {
	return s.items.Size()
}
