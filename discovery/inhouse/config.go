package inhouse

import (
	"errors"

	"github.com/bldsoft/gost/config"
	"github.com/bldsoft/gost/utils"
)

type Config struct {
	Embedded       bool                `mapstructure:"EMBEDDED" description:"If true, in-house discovery will use service bind address"`
	BindAddress    config.Address      `mapstructure:"BIND_ADDRESS" description:"For non embedded mode. Configuration related to what address to bind to and ports to listen on."`
	ClusterMembers []string            `mapstructure:"CLUSTER_MEMBERS" description:"Comma separated list of any existing member of the cluster to join it. Example: '127.0.0.1:3001'"`
	SecretKey      config.HiddenString `mapstructure:"SECRET_KEY" description:"SecretKey is used to encrypt messages. The value should be either 16, 24, or 32 bytes to select AES-128, AES-192, or AES-256."`
}

func (c *Config) SetDefaults() {
	c.Embedded = true
	c.BindAddress = "0.0.0.0:3001"
	c.SecretKey = "ZljFlK6atNj5U3VbHrDxRgFMHYcgEOpy"
}

func (c *Config) Validate() error {
	if !utils.IsIn(len(c.SecretKey), 16, 24, 32) {
		return errors.New("in-house secret key value should be either 16, 24, or 32 bytes")
	}
	return nil
}
