package channel

import (
	"context"
)

type Message struct {
	Data map[string]any
}

type Receiver interface {
	IsReceiver()
}

type Channel[R Receiver] interface {
	Send(ctx context.Context, receiver R, message Message) error
}
