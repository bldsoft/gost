package repository

import (
	"context"

	"github.com/bldsoft/gost/config"
	"github.com/golang-migrate/migrate/v4/source"
)

type IStorage interface {
	SetMigrationSrc(src source.Driver)
	Connect(server config.ConnectionString, database string)
	Disconnect(ctx context.Context) error
	IsReady() bool
}
