package config_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	gost "github.com/bldsoft/gost/config"
	"github.com/bldsoft/gost/examples/notes/pkg/config"
)

func newConfig(t *testing.T) config.Config {
	t.Helper()
	var cfg config.Config
	require.NoError(t, gost.SetDefaults(&cfg))
	cfg.AdminAPIKey = "test-key"

	return cfg
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*config.Config)
		wantErr string
	}{
		{name: "defaults with api key", mutate: func(*config.Config) {}},
		{name: "empty api key", mutate: func(c *config.Config) { c.AdminAPIKey = "" }, wantErr: "ADMIN_API_KEY"},
		{name: "zero purge interval", mutate: func(c *config.Config) { c.PurgeInterval = 0 }, wantErr: "NOTES_PURGE_INTERVAL"},
		{name: "negative purge after", mutate: func(c *config.Config) { c.PurgeAfter = -1 }, wantErr: "NOTES_PURGE_AFTER"},
		{name: "bad log level", mutate: func(c *config.Config) { c.Log.Level = "loud" }, wantErr: "loud"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := newConfig(t)
			tt.mutate(&cfg)
			err := gost.Validate(&cfg)
			if tt.wantErr == "" {
				require.NoError(t, err)

				return
			}
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestDefaults(t *testing.T) {
	cfg := newConfig(t)
	require.NoError(t, gost.Validate(&cfg))
	require.Equal(t, config.ServiceName, cfg.Server.ServiceName)
	require.Equal(t, "0.0.0.0:3000", cfg.Server.ServiceBindAddress.HostPort())
	require.Equal(t, config.ServiceName, cfg.Mongo.DbName)
	require.Equal(t, "info", cfg.Log.Level)
}

func TestConfigDescriptionGolden(t *testing.T) {
	var cfg config.Config
	require.NoError(t, gost.SetDefaults(&cfg))

	var buf bytes.Buffer
	require.NoError(t, gost.WriteConfigDescription(&cfg, "", gost.NewMarkdownFormatter(&buf, true)))

	path := filepath.Join("..", "..", "config.md")
	if os.Getenv("UPDATE_CONFIG_MD") != "" {
		require.NoError(t, os.WriteFile(path, buf.Bytes(), 0o644))

		return
	}
	want, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, string(want), buf.String(), "config.md is stale: run `make config-md` in examples/notes")
}
