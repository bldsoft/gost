package clickhouse

import (
	"net/url"

	"github.com/ClickHouse/clickhouse-go/v2"
	gost "github.com/bldsoft/gost/config"
)

type Config struct {
	options *clickhouse.Options
	Dsn     gost.ConnectionString `mapstructure:"DSN" description:"Clickhouse DSN"`
}

// SetDefaults ...
func (c *Config) SetDefaults() {
	c.Dsn = "clickhouse://127.0.0.1:9000/test?username=test_user&password=test_pass"
}

// Validate ...
func (c *Config) Validate() (err error) {
	if c.Dsn == "" {
		return nil
	}

	if err := c.prepareDsn(); err != nil {
		return err
	}

	c.options, err = clickhouse.ParseDSN(c.Dsn.String())
	return err
}

// convert v1 client DSN to v2
func (c *Config) prepareDsn() error {
	url, err := url.Parse(c.Dsn.String())
	if err != nil {
		return err
	}
	if url.Scheme != "clickhouse" {
		url.Scheme = "clickhouse"

		// remove default database from query and set it to path
		database := url.Query().Get("database")
		query := url.Query()
		query.Del("database")
		url.RawQuery = query.Encode()
		url.Path = "/" + database

		c.Dsn = gost.ConnectionString(url.String())
	}
	return nil
}
