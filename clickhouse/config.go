package clickhouse

import (
	"net/url"

	gost "github.com/bldsoft/gost/config"
)

type Config struct {
	Dsn gost.ConnectionString `mapstructure:"DSN"`
}

// SetDefaults ...
func (c *Config) SetDefaults() {
	c.Dsn = "tcp://127.0.0.1:9000?database=test&username=test_user&password=test_pass"
}

// Validate ...
func (c *Config) Validate() error {
	_, err := url.Parse(c.Dsn.String())
	return err
}
