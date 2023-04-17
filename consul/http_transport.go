package consul

import (
	"fmt"
	"net"
	"net/http"

	"github.com/hashicorp/consul/api"
)

func NewTransport(t http.RoundTripper, consulClient *api.Client) http.RoundTripper {
	return &transport{
		base:     t,
		resolver: NewResolver(consulClient),
	}
}

func NewTransportFromDiscovery(t http.RoundTripper, discovery *Discovery) http.RoundTripper {
	return NewTransport(t, discovery.ApiClient())
}

type transport struct {
	base     http.RoundTripper
	resolver *Resolver
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	host, _, err := net.SplitHostPort(req.URL.Host)
	if err != nil || host == "" {
		host = req.URL.Host
	}

	if net.ParseIP(host) != nil {
		return t.base.RoundTrip(req)
	}

	var addrs []string
	var res *http.Response

	if addrs, err = t.resolver.LookupServices(req.Context(), host); err != nil {
		return nil, err
	}

	if len(addrs) == 0 {
		return nil, fmt.Errorf("no addresses returned by the resolver for %s", host)
	}

	for _, addr := range addrs {
		if len(req.Host) == 0 {
			req.Host = req.URL.Host
		}
		req.URL.Host = addr
		res, err = t.base.RoundTrip(req)

		if err == nil || !isIdempotent(req.Method) {
			break
		}
	}

	return res, err
}

func isIdempotent(method string) bool {
	switch method {
	case "GET", "HEAD", "PUT", "DELETE", "OPTIONS":
		return true
	}
	return false
}
