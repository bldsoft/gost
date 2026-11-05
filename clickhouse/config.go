package clickhouse

import (
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/churl"

	gost "github.com/bldsoft/gost/config"
	"github.com/bldsoft/gost/utils"
)

type Config struct {
	Dsn     gost.ConnectionString `mapstructure:"DSN" description:"Clickhouse DSN"`
	options *clickhouse.Options
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

	if err = c.prepareDsn(); err != nil {
		return err
	}

	c.options, err = clickhouse.ParseDSN(c.Dsn.String())

	return err
}

// convert v1 client DSN to v2
func (c *Config) prepareDsn() error {
	url, err := churl.Parse(c.Dsn.String())
	if err != nil {
		return err
	}
	if !utils.IsIn(url.Scheme, "clickhouse") {
		if !utils.IsIn(url.Scheme, "http", "https") {
			url.Scheme = "clickhouse"
		}

		// v1 DSNs carry the database in the query; v2 reads it from the path. Move it only when
		// given, so a v2-style http://host/db keeps its database.
		if query := url.Query(); query.Has("database") {
			url.Path = "/" + query.Get("database")
			query.Del("database")
			url.RawQuery = query.Encode()
		}

		c.Dsn = gost.ConnectionString(url.String())
	}

	return nil
}
