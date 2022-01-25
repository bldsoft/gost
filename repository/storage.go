package repository

import (
	"context"

	"github.com/bldsoft/gost/config"
)

type IStorage interface {
	Connect(server config.ConnectionString, database string)
	Disconnect(ctx context.Context) error
	IsReady() bool
}
