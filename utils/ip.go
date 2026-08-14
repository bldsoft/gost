package utils

import (
	"errors"
	"net/netip"
	"strings"
)

const ipv4MappedPrefixBits = 96

var ErrInvalidIP = errors.New("invalid IP")

// CanonicalAddr returns the IPv4 form of an IPv4-mapped IPv6 address
// and strips any IPv6 zone.
func CanonicalAddr(addr netip.Addr) netip.Addr {
	return addr.Unmap().WithZone("")
}

func CanonicalPrefix(pfx netip.Prefix) (netip.Prefix, error) {
	if !pfx.IsValid() {
		return netip.Prefix{}, ErrInvalidIP
	}
	addr := CanonicalAddr(pfx.Addr())
	bits := pfx.Bits()
	if pfx.Addr().Is4In6() {
		bits -= ipv4MappedPrefixBits
		if bits < 0 || bits > addr.BitLen() {
			return netip.Prefix{}, ErrInvalidIP
		}
	}
	out := netip.PrefixFrom(addr, bits)
	if !out.IsValid() {
		return netip.Prefix{}, ErrInvalidIP
	}
	return out.Masked(), nil
}

func hostPrefix(addr netip.Addr) (netip.Prefix, error) {
	addr = CanonicalAddr(addr)
	if !addr.IsValid() {
		return netip.Prefix{}, ErrInvalidIP
	}
	pfx := netip.PrefixFrom(addr, addr.BitLen())
	if !pfx.IsValid() {
		return netip.Prefix{}, ErrInvalidIP
	}
	return pfx, nil
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
	return hostPrefix(addr)
}
