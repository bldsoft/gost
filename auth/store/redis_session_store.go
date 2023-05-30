package store

import (
	"context"
	"strings"

	"github.com/go-redis/redis"
	"github.com/gorilla/sessions"
	"github.com/rbcervilla/redisstore"
)

type RedisSessionStore struct {
	*redisstore.RedisStore
	client     redis.UniversalClient
	keyPrefix  string
	serializer redisstore.SessionSerializer
}

func NewRedisStore(client redis.UniversalClient) (*RedisSessionStore, error) {
	store, err := redisstore.NewRedisStore(client)
	if err != nil {
		return nil, err
	}
	keyPrefix := "session:"
	store.KeyPrefix(keyPrefix)
	serializer := &redisstore.GobSerializer{}
	store.Serializer(serializer)
	return &RedisSessionStore{RedisStore: store, client: client, keyPrefix: keyPrefix, serializer: serializer}, nil
}

func (s *RedisSessionStore) AllSessions(ctx context.Context, name string, offset, limit int) ([]*sessions.Session, error) {
	// TODO: optimize
	keys, err := s.client.Keys(s.keyPrefix + "*").Result()
	if err != nil {
		return nil, err
	}

	if offset > len(keys) {
		return nil, nil
	}
	limit += offset
	if limit > len(keys) {
		limit = len(keys)
	}
	keys = keys[offset:limit]

	res := make([]*sessions.Session, 0, len(keys))
	for _, key := range keys {
		session := sessions.NewSession(s, name)

		cmd := s.client.Get(key)
		if cmd.Err() != nil {
			return nil, cmd.Err()
		}

		b, err := cmd.Bytes()
		if err != nil {
			return nil, err
		}

		if err := s.serializer.Deserialize(b, session); err != nil {
			return nil, err
		}

		session.ID = strings.TrimPrefix(key, s.keyPrefix)
		res = append(res, session)
	}

	return res, nil
}

func (s *RedisSessionStore) KillSession(ctx context.Context, sessionID string) error {
	return s.client.Del(s.keyPrefix + sessionID).Err()
}
