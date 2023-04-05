package consul

import (
	"context"
	"net"

	"github.com/hashicorp/consul/api"
)

func GrpcDialer(consulClient *api.Client) func(ctx context.Context, s string) (net.Conn, error) {
	dialer := DefaultDialer(consulClient)
	return func(ctx context.Context, s string) (net.Conn, error) {
		return dialer.DialContext(ctx, "tcp", s)
	}
}

func GrpcDialerFromDiscovery(d *Discovery) func(ctx context.Context, s string) (net.Conn, error) {
	return GrpcDialer(d.ApiClient())
}
