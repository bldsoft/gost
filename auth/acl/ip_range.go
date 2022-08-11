package acl

import (
	"errors"
	"net"
	"strings"
)

type IpRange struct {
	Ip   []net.IP
	Cidr []*net.IPNet
}

func IpRangeFromStrings(strs []string) (*IpRange, error) {
	var ipRange IpRange
	for _, s := range strs {
		if strings.Contains(s, "/") {
			_, network, err := net.ParseCIDR(s)
			if err != nil {
				return nil, err
			}
			ipRange.Cidr = append(ipRange.Cidr, network)
		} else {
			ip := net.ParseIP(s)
			if ip == nil {
				return nil, errors.New("unable to parse IP address")
			}
			ipRange.Ip = append(ipRange.Ip, ip)
		}
	}
	return &ipRange, nil
}

func (r IpRange) Empty() bool {
	return len(r.Ip) == 0 && len(r.Cidr) == 0
}

func (r IpRange) isInIPs(client net.IP, ips []net.IP) bool {
	for _, ip := range ips {
		if client.Equal(ip) {
			return true
		}
	}
	return false
}

func (r IpRange) isInSubnets(ip net.IP, subs []*net.IPNet) bool {
	for _, subnet := range subs {
		if subnet.Contains(ip) {
			return true
		}
	}
	return false
}

func (r IpRange) Contains(ip net.IP) bool {
	return r.isInSubnets(ip, r.Cidr) || r.isInIPs(ip, r.Ip)
}
