package acl

import (
	"errors"
	"net"
	"net/http"

	"github.com/bldsoft/gost/controller"
)

type Config struct {
	Allow []string `mapstructure:"ACL_ALLOW" description:"If not empty it allows access only for the specified networks or addresses. Example: \"192.168.1.1,10.1.1.0/16\""`
	allow *IpRange
	Deny  []string `mapstructure:"ACL_DENY" description:"Denies access for the specified networks or addresses. Example: \"192.168.1.1,10.1.1.0/16\""`
	deny  *IpRange
}

func (c *Config) SetDefaults() {}

func (c *Config) Validate() (err error) {
	c.allow, err = IpRangeFromStrings(c.Allow)
	if err != nil {
		return err
	}

	c.deny, err = IpRangeFromStrings(c.Deny)
	if err != nil {
		return err
	}

	return nil
}

type Acl struct {
	controller controller.BaseController

	Allow *IpRange
	Deny  *IpRange
}

func MiddlewareFromConfig(cfg Config) func(next http.Handler) http.Handler {
	return Acl{Allow: cfg.allow, Deny: cfg.deny}.Middleware
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
