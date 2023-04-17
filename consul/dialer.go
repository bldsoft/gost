package consul

import (
	"context"
	"net"
	"time"

	"github.com/hashicorp/consul/api"
)

type Dialer struct {
	base     *net.Dialer
	resolver *Resolver
}

func DefaultDialer(consulClient *api.Client) *Dialer {
	resolver := NewResolver(consulClient)
	baseDialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	return NewDialer(baseDialer, resolver)
}

func DefaultDialerFromDiscovery(d *Discovery) *Dialer {
	return DefaultDialer(d.consulClient)
}

func NewDialer(d *net.Dialer, r *Resolver) *Dialer {
	return &Dialer{d, r}
}

func (d *Dialer) Dial(network, address string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, address)
}

func (d *Dialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil || host == "" {
		host = addr
	}

	if net.ParseIP(host) != nil {
		return d.base.DialContext(ctx, network, addr)
	}

	addrs, err := d.resolver.LookupServices(ctx, host)
	if err != nil {
		return nil, err
	}
	for _, addr := range addrs {
		var conn net.Conn
		conn, err = d.base.DialContext(ctx, network, addr)
		if err == nil {
			return conn, nil
		}
	}
	return nil, err
}
