package mongotest

import (
	"context"

	"github.com/bldsoft/gost/config"
	"github.com/bldsoft/gost/internal/testcontainers"
	"github.com/bldsoft/gost/mongo"
)

func Start(ctx context.Context) (mongo.Config, func(), error) {
	uri, cleanup, err := testcontainers.StartMongo(ctx)
	if err != nil {
		return mongo.Config{}, cleanup, err
	}

	var cfg mongo.Config
	cfg.SetDefaults()
	cfg.Server = config.ConnectionString(uri)

	return cfg, cleanup, nil
}
