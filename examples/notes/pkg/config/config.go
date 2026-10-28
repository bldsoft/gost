package config

import (
	"errors"
	"time"

	"github.com/bldsoft/gost/config"
	"github.com/bldsoft/gost/discovery/common"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/mongo"
	"github.com/bldsoft/gost/server"
)

const ServiceName = "notes"

type Config struct {
	Server    server.Config
	Log       log.Config
	Mongo     mongo.Config
	Discovery common.Config `mapstructure:"DISCOVERY"`

	AdminAPIKey   config.HiddenString `mapstructure:"ADMIN_API_KEY" description:"API key expected in the X-API-Key header by /notes/env. Required: the service refuses to start without it."`
	PurgeInterval time.Duration       `mapstructure:"NOTES_PURGE_INTERVAL" description:"How often soft-deleted notes are purged (Go duration, e.g. 1h)"`
	PurgeAfter    time.Duration       `mapstructure:"NOTES_PURGE_AFTER" description:"Soft-deleted notes older than this are removed permanently (Go duration, e.g. 720h)"`
}

func (c *Config) SetDefaults() {
	c.Server.ServiceName = ServiceName
	c.Server.ServiceBindAddress = "0.0.0.0:3000"
	c.Mongo.DbName = ServiceName
	c.PurgeInterval = time.Hour
	c.PurgeAfter = 30 * 24 * time.Hour
}

func (c *Config) Validate() error {
	if c.AdminAPIKey == "" {
		return errors.New("ADMIN_API_KEY is required")
	}
	if c.PurgeInterval <= 0 {
		return errors.New("NOTES_PURGE_INTERVAL must be positive")
	}
	if c.PurgeAfter <= 0 {
		return errors.New("NOTES_PURGE_AFTER must be positive")
	}

	return nil
}
