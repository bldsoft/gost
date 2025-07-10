package aerospike

import (
	aero "github.com/aerospike/aerospike-client-go/v8"
)

type Storage struct {
	*aero.Client
	keyPrefix string
	namespace string
}

func NewStorage(cfg Config) (*Storage, error) {
	client, err := aero.NewClient(cfg.Host, cfg.Port)
	if err != nil {
		return nil, err
	}
	return &Storage{
		Client:    client,
		keyPrefix: cfg.KeyPrefix,
	}, nil
}
