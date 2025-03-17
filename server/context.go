package server

import (
	"context"
	"net"
)

func ConnFromContext(ctx context.Context) net.Conn {
	if conn, ok := ctx.Value(connContextKey).(net.Conn); ok {
		return conn
	}
	return nil
}
