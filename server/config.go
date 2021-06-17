package server

import "fmt"

type Config struct {
	Host string `mapstructure:"SERVICE_HOST"`
	Port int    `mapstructure:"SERVICE_PORT"`
}

func (c *Config) ServiceAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// SetDefaults ...
func (c *Config) SetDefaults() {
	c.Host = "0.0.0.0"
	c.Port = 3000
}

// Validate ...
func (c *Config) Validate() error {
	var err error
	return err
}
