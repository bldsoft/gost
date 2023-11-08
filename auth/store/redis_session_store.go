package store

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/gob"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bldsoft/gost/log"
	"github.com/go-redis/redis"
	"github.com/gorilla/sessions"
)

type RedisSessionStore struct {
	options sessions.Options
	// key prefix with which the session will be stored
	keyPrefix string
	// key generator
	keyGen KeyGenFunc
	// session serializer
	serializer SessionSerializer

	client redis.UniversalClient
}

// KeyGenFunc defines a function used by store to generate a key
type KeyGenFunc func() (string, error)

func NewRedisStore(client redis.UniversalClient) (*RedisSessionStore, error) {
	return &RedisSessionStore{
		options: sessions.Options{
			Path:   "/",
			MaxAge: 86400 * 30,
		},
		client:     client,
		keyPrefix:  "session:",
		keyGen:     generateRandomKey,
		serializer: GobSerializer{},
	}, nil
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

// Get returns a session for the given name after adding it to the registry.
func (s *RedisSessionStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

// New returns a session for the given name without adding it to the registry.
func (s *RedisSessionStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(s, name)
	opts := s.options
	session.Options = &opts
	session.IsNew = true

	c, err := r.Cookie(name)
	if err != nil {
		return session, nil
	}
	session.ID = c.Value

	err = s.load(session)
	if err == nil {
		session.IsNew = false
	} else if err == redis.Nil {
		err = nil // no data stored
	}
	return session, err
}

// Save adds a single session to the response.
//
// If the Options.MaxAge of the session is <= 0 then the session file will be
// deleted from the store. With this process it enforces the properly
// session cookie handling so no need to trust in the cookie management in the
// web browser.
func (s *RedisSessionStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	// Delete if max-age is <= 0
	if session.Options.MaxAge <= 0 {
		if err := s.delete(session); err != nil {
			return err
		}
		http.SetCookie(w, sessions.NewCookie(session.Name(), "", session.Options))
		return nil
	}

	if session.ID == "" {
		id, err := s.keyGen()
		if err != nil {
			return errors.New("redisstore: failed to generate session id")
		}
		session.ID = id
	}
	if err := s.save(session); err != nil {
		return err
	}

	http.SetCookie(w, sessions.NewCookie(session.Name(), session.ID, session.Options))
	return nil
}

// Options set options to use when a new session is created
func (s *RedisSessionStore) Options(opts sessions.Options) {
	s.options = opts
}

// KeyPrefix sets the key prefix to store session in Redis
func (s *RedisSessionStore) KeyPrefix(keyPrefix string) {
	s.keyPrefix = keyPrefix
}

// KeyGen sets the key generator function
func (s *RedisSessionStore) KeyGen(f KeyGenFunc) {
	s.keyGen = f
}

// Serializer sets the session serializer to store session
func (s *RedisSessionStore) Serializer(ss SessionSerializer) {
	s.serializer = ss
}

// save writes session in Redis
func (s *RedisSessionStore) save(session *sessions.Session) error {
	b, err := s.serializer.Serialize(session)
	if err != nil {
		return err
	}

	if session.IsNew {
		return s.client.Set(s.keyPrefix+session.ID, b, time.Duration(session.Options.MaxAge)*time.Second).Err()
	}
	return s.client.SetXX(s.keyPrefix+session.ID, b, time.Duration(session.Options.MaxAge)*time.Second).Err()
}

// load reads session from Redis
func (s *RedisSessionStore) load(session *sessions.Session) error {
	cmd := s.client.Get(s.keyPrefix + session.ID)
	if cmd.Err() != nil {
		return cmd.Err()
	}

	b, err := cmd.Bytes()
	if err != nil {
		return err
	}

	return s.serializer.Deserialize(b, session)
}

// delete deletes session in Redis
func (s *RedisSessionStore) delete(session *sessions.Session) error {
	return s.client.Del(s.keyPrefix + session.ID).Err()
}

func (s *RedisSessionStore) KillSessions(ctx context.Context, sessionIDs ...string) error {
	for i, id := range sessionIDs {
		sessionIDs[i] = s.keyPrefix + id
	}
	return s.client.Del(sessionIDs...).Err()
}

func (s *RedisSessionStore) KillUserSessions(_ context.Context, _ string) error {
	log.Error("kill user sessions is not implemented for redis")
	return nil
}

// SessionSerializer provides an interface for serialize/deserialize a session
type SessionSerializer interface {
	Serialize(s *sessions.Session) ([]byte, error)
	Deserialize(b []byte, s *sessions.Session) error
}

// Gob serializer
type GobSerializer struct{}

func (gs GobSerializer) Serialize(s *sessions.Session) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(s.Values)
	if err == nil {
		return buf.Bytes(), nil
	}
	return nil, err
}

func (gs GobSerializer) Deserialize(d []byte, s *sessions.Session) error {
	dec := gob.NewDecoder(bytes.NewBuffer(d))
	return dec.Decode(&s.Values)
}

// generateRandomKey returns a new random key
func generateRandomKey() (string, error) {
	k := make([]byte, 64)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		return "", err
	}
	return strings.TrimRight(base32.StdEncoding.EncodeToString(k), "="), nil
}
