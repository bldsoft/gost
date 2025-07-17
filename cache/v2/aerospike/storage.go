package aerospike

import (
	"fmt"
	"time"

	aero "github.com/aerospike/aerospike-client-go/v8"
	"github.com/bldsoft/gost/log"

	logger "github.com/aerospike/aerospike-client-go/v8/logger"
	"github.com/mitchellh/mapstructure"
)

type Storage struct {
	*aero.Client
	namespace string
}

func NewStorage(cfg Config) (*Storage, error) {
	logger.Logger.SetLevel(logger.DEBUG)

	policy := aero.NewClientPolicy()
	policy.ConnectionQueueSize = 2000
	policy.Timeout = 500 * time.Millisecond
	policy.IdleTimeout = 30 * time.Second
	policy.FailIfNotConnected = false

	client, err := aero.NewClientWithPolicy(policy, cfg.Host, cfg.Port)
	if err != nil {
		return nil, err
	}
	if _, err := client.WarmUp(0); err != nil {
		log.WarnWithFields(log.Fields{"err": err}, "failed to warm up aerospike client")
	}
	return &Storage{
		Client:    client,
		namespace: cfg.Namespace,
	}, nil
}

func (s *Storage) Stat() (*Stats, error) {
	stats, err := s.Stats()
	if err != nil {
		return nil, err
	}
	var stat Stats
	if err := mapstructure.WeakDecode(stats, &stat); err != nil {
		return nil, fmt.Errorf("failed to decode aerospike stats: %w", err)
	}
	return &stat, nil
}
