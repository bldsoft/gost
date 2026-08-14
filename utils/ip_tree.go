package utils

import (
	"net/netip"

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
		pfx, err := IPKeyToPrefix(ipCidr)
		if err != nil {
			continue
		}
		s.items.Insert(pfx)
	}

	return s
}

func (s *IPTreeSet) Put(ipCidrs ...string) error {
	for _, ipCidr := range ipCidrs {
		pfx, err := IPKeyToPrefix(ipCidr)
		if err != nil {
			return err
		}
		s.items.Insert(pfx)
	}
	return nil
}

func (s *IPTreeSet) Delete(ipCidrs ...string) error {
	for _, ipCidr := range ipCidrs {
		pfx, err := IPKeyToPrefix(ipCidr)
		if err != nil {
			return err
		}
		s.items.Delete(pfx)
	}
	return nil
}

func (s *IPTreeSet) Match(ip netip.Addr) bool {
	return s.items.Lookup(CanonicalAddr(ip))
}

func (s *IPTreeSet) Len() int {
	return s.items.Size()
}

type IPTree[V any] struct {
	table bart.Table[V]
}

func NewIPTree[V any]() *IPTree[V] {
	return &IPTree[V]{}
}

func (t *IPTree[V]) Insert(ipCidr string, val V) error {
	pfx, err := IPKeyToPrefix(ipCidr)
	if err != nil {
		return err
	}
	t.table.Insert(pfx, val)
	return nil
}

func (t *IPTree[V]) Delete(ipCidr string) error {
	pfx, err := IPKeyToPrefix(ipCidr)
	if err != nil {
		return err
	}
	t.table.Delete(pfx)
	return nil
}

func (t *IPTree[V]) Lookup(ip netip.Addr) (V, bool) {
	return t.table.Lookup(CanonicalAddr(ip))
}

func (t *IPTree[V]) LookupString(ip string) (V, bool) {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		var zero V
		return zero, false
	}
	return t.Lookup(addr)
}
