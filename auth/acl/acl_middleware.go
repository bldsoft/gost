package acl

import (
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/bldsoft/gost/controller"
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

type Acl struct {
	controller controller.BaseController

	Allow *IpRange
	Deny  *IpRange
}

func (m Acl) getIP(r *http.Request) (net.IP, error) {
	var ip string
	var err error

	ip, _, err = net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return nil, err
	}

	ret := net.ParseIP(ip)
	if ret == nil {
		return nil, errors.New("acl: unable to parse address")
	}

	return ret, nil
}

func (m Acl) Middleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ip, err := m.getIP(r)
		if err != nil {
			m.controller.ResponseError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if m.Deny != nil && m.Deny.Contains(ip) {
			m.controller.ResponseError(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		if m.Allow != nil && !m.Allow.Empty() && !m.Allow.Contains(ip) {
			m.controller.ResponseError(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
