package storage

import (
	"context"
)

type IStorage interface {
	Connect() error
	Disconnect(ctx context.Context) error
	IsReady() bool
}
