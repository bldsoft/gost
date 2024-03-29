package storage

import (
	"context"
)

type IStorage interface {
	Connect()
	Disconnect(ctx context.Context) error
	IsReady() bool
}
