package mongo

import (
	"github.com/bldsoft/gost/config"
)

type Config struct {
	Server config.ConnectionString `mapstructure:"DATABASE_SERVER"`
	DBName string                  `mapstructure:"DATABASE_DBNAME"`
}

// SetDefaults ...
func (c *Config) SetDefaults() {
	c.Server = "mongodb://localhost:27017"
	c.DBName = "test"
}

// Validate ...
func (c *Config) Validate() error {
	return nil
}
