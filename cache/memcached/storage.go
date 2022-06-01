package memcached

import (
	"context"
	"fmt"

	"github.com/bldsoft/gost/log"
	"github.com/bradfitz/gomemcache/memcache"
	stat_client "github.com/grobie/gomemcache/memcache"
	"github.com/mitchellh/mapstructure"
)

type Storage struct {
	*memcache.Client
	statClient *stat_client.Client
}

func NewStorage(cfg Config) *Storage {
	client := memcache.New(cfg.Servers...)
	if err := client.Ping(); err != nil {
		log.Fatalf("Memcached connection failed: %v", err)
	}
	statClient, err := stat_client.New(cfg.Servers...)
	if err != nil {
		log.Logger.WarnWithFields(log.Fields{"err": err}, "failed to create stat memcached client")
	}
	return &Storage{
		Client:     client,
		statClient: statClient,
	}
}

func (s *Storage) Stats(ctx context.Context) ([]*Stats, error) {
	stats, err := s.statClient.Stats()
	if err != nil {
		return nil, err
	}
	res := make([]*Stats, 0, len(stats))
	for key, s := range stats {
		var stat Stats
		if err := mapstructure.WeakDecode(s.Stats, &stat); err != nil {
			return nil, fmt.Errorf("failed to decode memcached stats: %w", err)
		}
		stat.Instance = key.String()
		res = append(res, &stat)
	}
	return res, nil
}
