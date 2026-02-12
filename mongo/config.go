package mongo

import (
	"github.com/bldsoft/gost/config"
	mm "github.com/golang-migrate/migrate/v4/database/mongodb"
)

type Config struct {
	Server              config.ConnectionString `mapstructure:"DATABASE_SERVER" description:"MongoDB connection string"`
	DbName              string                  `mapstructure:"DATABASE_DBNAME" description:"Service database name"`
	MigrationCollection string                  `mapstructure:"DATABASE_MIGRATION_COLLECTION" description:"MongoDB migration collection name"`
}

// SetDefaults ...
func (c *Config) SetDefaults() {
	c.Server = "mongodb://localhost:27017"
	c.DbName = "test"
	c.MigrationCollection = mm.DefaultMigrationsCollection
}

// Validate ...
func (c *Config) Validate() error {
	return nil
}
