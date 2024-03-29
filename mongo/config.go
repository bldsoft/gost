package mongo

import (
	"github.com/bldsoft/gost/config"
)

type Config struct {
	Server config.ConnectionString `mapstructure:"DATABASE_SERVER" description:"MongoDB connection string"`
	DbName string                  `mapstructure:"DATABASE_DBNAME" description:"Service database name"`
}

// SetDefaults ...
func (c *Config) SetDefaults() {
	c.Server = "mongodb://localhost:27017"
	c.DbName = "test"
}

// Validate ...
func (c *Config) Validate() error {
	return nil
}
