package utils

import (
	"errors"
	"net/netip"
	"strings"
)

const ipv4MappedPrefixBits = 96

var ErrInvalidIP = errors.New("invalid IP")

// CanonicalAddr returns the IPv4 form of an IPv4-mapped IPv6 address.
func CanonicalAddr(addr netip.Addr) netip.Addr {
	if !addr.IsValid() {
		return addr
	}
	return addr.Unmap()
}

func CanonicalPrefix(pfx netip.Prefix) (netip.Prefix, error) {
	if !pfx.IsValid() {
		return netip.Prefix{}, ErrInvalidIP
	}
	addr := pfx.Addr().Unmap()
	bits := pfx.Bits()
	if pfx.Addr().Is4In6() {
		bits -= ipv4MappedPrefixBits
		if bits < 0 || bits > addr.BitLen() {
			return netip.Prefix{}, ErrInvalidIP
		}
	}
	return netip.PrefixFrom(addr, bits), nil
}

func IPKeyToPrefix(ip string) (netip.Prefix, error) {
	if strings.Contains(ip, "/") {
		pfx, err := netip.ParsePrefix(ip)
		if err != nil {
			return netip.Prefix{}, err
		}
		return CanonicalPrefix(pfx)
	}
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return netip.Prefix{}, err
	}
	addr = CanonicalAddr(addr)
	return netip.PrefixFrom(addr, addr.BitLen()), nil
}
