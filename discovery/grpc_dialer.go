package discovery

import (
	"context"
	"net"
)

func GrpcDialer(d Discovery) func(ctx context.Context, s string) (net.Conn, error) {
	dialer := DefaultDialer(d)
	return func(ctx context.Context, s string) (net.Conn, error) {
		return dialer.DialContext(ctx, "tcp", s)
	}
}

func GrpcDialerFromDiscovery(d Discovery) func(ctx context.Context, s string) (net.Conn, error) {
	return GrpcDialer(d)
}
