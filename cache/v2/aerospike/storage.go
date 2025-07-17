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
	cfg       Config
}

func NewStorage(cfg Config) (*Storage, error) {
	logger.Logger.SetLevel(logger.DEBUG)

	policy := aero.NewClientPolicy()
	policy.ConnectionQueueSize = cfg.ConnectionPolicy.ConnectionQueueSize
	policy.Timeout = time.Duration(cfg.ConnectionPolicy.TimeOutMs) * time.Millisecond
	policy.IdleTimeout = time.Duration(cfg.ConnectionPolicy.IdleTimeoutMs) * time.Millisecond
	policy.FailIfNotConnected = cfg.ConnectionPolicy.FailIfNotConnected

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
		cfg:       cfg,
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

func (s *Storage) getWritePolicy(generation uint32, expiration uint32) *aero.WritePolicy {
	wp := aero.NewWritePolicy(0, 0)
	wp.TotalTimeout = time.Duration(s.cfg.WritePolicy.TotalTimeoutMs) * time.Millisecond
	wp.MaxRetries = s.cfg.WritePolicy.MaxRetries
	wp.SleepBetweenRetries = time.Duration(s.cfg.WritePolicy.SleepBetweenRetriesMs) * time.Millisecond
	wp.SocketTimeout = time.Duration(s.cfg.WritePolicy.SocketTimeoutMs) * time.Millisecond
	return wp
}

func (s *Storage) getReadPolicy() *aero.BasePolicy {
	rp := aero.NewPolicy()
	rp.TotalTimeout = time.Duration(s.cfg.ReadPolicy.TotalTimeoutMs) * time.Millisecond
	rp.MaxRetries = s.cfg.ReadPolicy.MaxRetries
	rp.SleepBetweenRetries = time.Duration(s.cfg.ReadPolicy.SleepBetweenRetriesMs) * time.Millisecond
	rp.SocketTimeout = time.Duration(s.cfg.ReadPolicy.SocketTimeoutMs) * time.Millisecond
	return rp
}
