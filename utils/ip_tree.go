package utils

import (
	"net/netip"
	"strings"

	"github.com/gaissmai/bart"
)

type IPTreeSet struct {
	items *bart.Lite
}

func NewIPTreeSet() *IPTreeSet {
	return &IPTreeSet{
		items: new(bart.Lite),
	}
}

func (s *IPTreeSet) ensure() {
	if s.items == nil {
		s.items = new(bart.Lite)
	}
}

func (s *IPTreeSet) Put(ipCidrs ...string) error {
	for _, ipCidr := range ipCidrs {
		if strings.Contains(ipCidr, "/") {
			pfx, err := netip.ParsePrefix(ipCidr)
			if err != nil {
				return err
			}
			if err := s.PutPrefixes(pfx); err != nil {
				return err
			}
			continue
		}
		addr, err := netip.ParseAddr(ipCidr)
		if err != nil {
			return err
		}
		if err := s.PutIPs(addr); err != nil {
			return err
		}
	}
	return nil
}

func (s *IPTreeSet) PutIPs(ips ...netip.Addr) error {
	s.ensure()
	for _, ip := range ips {
		pfx, err := hostPrefix(ip)
		if err != nil {
			return err
		}
		s.items.Insert(pfx)
	}
	return nil
}

func (s *IPTreeSet) PutPrefixes(pfxs ...netip.Prefix) error {
	s.ensure()
	for _, pfx := range pfxs {
		canon, err := CanonicalPrefix(pfx)
		if err != nil {
			return err
		}
		s.items.Insert(canon)
	}
	return nil
}

func (s *IPTreeSet) Delete(ipCidrs ...string) error {
	if s == nil || s.items == nil {
		return nil
	}
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
	if s == nil || s.items == nil {
		return false
	}
	return s.items.Lookup(CanonicalAddr(ip))
}

func (s *IPTreeSet) Len() int {
	if s == nil || s.items == nil {
		return 0
	}
	return s.items.Size()
}

type IPTree[V any] struct {
	table *bart.Table[V]
}

func NewIPTree[V any]() *IPTree[V] {
	return &IPTree[V]{
		table: new(bart.Table[V]),
	}
}

func (t *IPTree[V]) ensure() {
	if t.table == nil {
		t.table = new(bart.Table[V])
	}
}

func (t *IPTree[V]) Insert(ipCidr string, val V) error {
	if strings.Contains(ipCidr, "/") {
		pfx, err := netip.ParsePrefix(ipCidr)
		if err != nil {
			return err
		}
		return t.PutPrefixes(val, pfx)
	}
	addr, err := netip.ParseAddr(ipCidr)
	if err != nil {
		return err
	}
	return t.PutIPs(val, addr)
}

func (t *IPTree[V]) PutIPs(val V, ips ...netip.Addr) error {
	t.ensure()
	for _, ip := range ips {
		pfx, err := hostPrefix(ip)
		if err != nil {
			return err
		}
		t.table.Insert(pfx, val)
	}
	return nil
}

func (t *IPTree[V]) PutPrefixes(val V, pfxs ...netip.Prefix) error {
	t.ensure()
	for _, pfx := range pfxs {
		canon, err := CanonicalPrefix(pfx)
		if err != nil {
			return err
		}
		t.table.Insert(canon, val)
	}
	return nil
}

func (t *IPTree[V]) Delete(ipCidr string) error {
	if t == nil || t.table == nil {
		return nil
	}
	pfx, err := IPKeyToPrefix(ipCidr)
	if err != nil {
		return err
	}
	t.table.Delete(pfx)
	return nil
}

func (t *IPTree[V]) Lookup(ip netip.Addr) (V, bool) {
	var zero V
	if t == nil || t.table == nil {
		return zero, false
	}
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
