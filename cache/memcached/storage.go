package memcached

import (
	"github.com/bldsoft/gost/log"
	"github.com/bradfitz/gomemcache/memcache"
)

type Storage struct {
	*memcache.Client
}

func NewStorage(cfg Config) *Storage {
	client := memcache.New(cfg.Servers...)
	if err := client.Ping(); err != nil {
		log.Fatalf("Memcached connection failed: %v", err)
	}
	return &Storage{
		Client: client,
	}
}
