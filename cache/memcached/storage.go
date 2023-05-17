package memcached

import (
	"context"
	"fmt"
	"time"

	"github.com/bldsoft/gost/log"
	"github.com/bradfitz/gomemcache/memcache"
	stat_client "github.com/grobie/gomemcache/memcache"
	"github.com/mitchellh/mapstructure"
)

type Storage struct {
	*memcache.Client
	statClient *stat_client.Client
	keyPrefix  string
}

func NewStorage(cfg Config) (*Storage, *Storage) {
	master := singleStorage(cfg.Servers, cfg.TimeoutMs, cfg.KeyPrefix)
	if len(cfg.ReadOnlyServers) == 0 {
		return master, nil
	}
	return master, singleStorage(cfg.ReadOnlyServers, cfg.TimeoutMs, cfg.KeyPrefix)
}

func singleStorage(servers []string, timeOut int, keyPrefix string) *Storage {
	client := memcache.New(servers...)
	if timeOut != 0 {
		client.Timeout = time.Duration(timeOut) * time.Millisecond
	}
	if err := client.Ping(); err != nil {
		log.Panicf("Memcached connection failed: %v", err)
	}
	statClient, err := stat_client.New(servers...)
	if err != nil {
		log.Logger.WarnWithFields(log.Fields{"err": err}, "failed to create stat memcached client")
	}
	return &Storage{
		Client:     client,
		statClient: statClient,
		keyPrefix:  keyPrefix,
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

func (s *Storage) PrepareKey(key string) string {
	if len(s.keyPrefix) == 0 {
		return key
	}
	return s.keyPrefix + key
}

func (s *Storage) PrepareKeys(keys []string) []string {
	if len(s.keyPrefix) == 0 {
		return keys
	}

	newKeys := make([]string, len(keys))
	for i, k := range keys {
		newKeys[i] = s.PrepareKey(k)
	}
	return newKeys
}
