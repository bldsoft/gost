package server

import "fmt"

type Config struct {
	Name string `mapstructure:"SERVICE_NAME"`
	Host string `mapstructure:"SERVICE_HOST"`
	Port int    `mapstructure:"SERVICE_PORT"`
}

func (c *Config) ServiceAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// SetDefaults ...
func (c *Config) SetDefaults() {
	c.Name = "default_name"
	c.Host = "0.0.0.0"
	c.Port = 3000
}

// Validate ...
func (c *Config) Validate() error {
	var err error
	return err
}
