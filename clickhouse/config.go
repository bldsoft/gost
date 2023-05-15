package clickhouse

import (
	"github.com/ClickHouse/clickhouse-go/v2"
	gost "github.com/bldsoft/gost/config"
)

type Config struct {
	Dsn     gost.ConnectionString `mapstructure:"DSN" description:"Clickhouse DSN"`
	options *clickhouse.Options
}

// SetDefaults ...
func (c *Config) SetDefaults() {
	c.Dsn = "tcp://127.0.0.1:9000/test?username=test_user&password=test_pass"
}

// Validate ...
func (c *Config) Validate() (err error) {
	c.options, err = clickhouse.ParseDSN(c.Dsn.String())
	return err
}
