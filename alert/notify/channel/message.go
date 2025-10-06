package channel

import (
	"context"
)

type Message struct {
	Data map[string]any
}

type Receiver any

type Channel[R Receiver] interface {
	Send(ctx context.Context, receiver R, message Message) error
}
