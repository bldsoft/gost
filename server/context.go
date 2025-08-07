package server

import (
	"context"
	"net"
)

func ConnFromContext(ctx context.Context) net.Conn {
	conn, _ := ctx.Value(connContextKey{}).(net.Conn)
	return conn
}
