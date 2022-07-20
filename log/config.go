package log

import (
	"github.com/rs/zerolog"
)

type Config struct {
	File     string `mapstructure:"LOG_FILE" description:"-"`
	Level    string `mapstructure:"LOG_LEVEL" description:"Log level"`
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
