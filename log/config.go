package log

import (
	"github.com/rs/zerolog"
)

type Config struct {
	File     string `mapstructure:"LOG_FILE"`
	Level    string `mapstructure:"LOG_LEVEL"`
	Exporter LogExporterConfig
}

// SetDefaults ...
func (c *Config) SetDefaults() {
	c.Level = zerolog.InfoLevel.String()
}

// Validate ...
func (c *Config) Validate() error {
	_, err := zerolog.ParseLevel(c.Level)
	return err
}
