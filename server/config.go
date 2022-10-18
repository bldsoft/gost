package server

import "fmt"

type Config struct {
	Name string `mapstructure:"SERVICE_NAME" description:"Unique service instance name"`
	Host string `mapstructure:"SERVICE_HOST" description:"IP address, or a host name that can be resolved to IP addresses"`
	Port int    `mapstructure:"SERVICE_PORT" description:"Service port"`
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
