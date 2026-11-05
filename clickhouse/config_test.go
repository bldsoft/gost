package clickhouse

import (
	"testing"

	"github.com/stretchr/testify/require"

	gost "github.com/bldsoft/gost/config"
)

func TestConfigValidate_DSN(t *testing.T) {
	tests := []struct {
		name     string
		dsn      string
		addr     []string
		database string
	}{
		{"native keeps the path", "clickhouse://u:p@h1:9000/db", []string{"h1:9000"}, "db"},
		{"http keeps the path", "http://u:p@h1:8123/db", []string{"h1:8123"}, "db"},
		{"http with several hosts keeps the path", "http://u:p@h1:8123,h2:8123,h3:8123/db", []string{"h1:8123", "h2:8123", "h3:8123"}, "db"},
		{"v1 database query moves to the path", "http://u:p@h1:8123?database=db", []string{"h1:8123"}, "db"},
		{"v1 database query wins over the path", "http://u:p@h1:8123/other?database=db", []string{"h1:8123"}, "db"},
		{"tcp scheme becomes clickhouse", "tcp://u:p@h1:9000?database=db", []string{"h1:9000"}, "db"},
		{"no database at all", "http://u:p@h1:8123", []string{"h1:8123"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{Dsn: gost.ConnectionString(tt.dsn)}
			require.NoError(t, cfg.Validate())
			require.Equal(t, tt.addr, cfg.options.Addr)
			require.Equal(t, tt.database, cfg.options.Auth.Database)
			require.Equal(t, "u", cfg.options.Auth.Username)
			require.Equal(t, "p", cfg.options.Auth.Password)
		})
	}
}
