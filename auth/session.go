package auth

import (
	"sync"

	"github.com/gorilla/sessions"
)

type Session struct {
	session *sessions.Session
	mtx     sync.RWMutex
}

func wrapSession(s *sessions.Session) *Session {
	return &Session{session: s}
}

func (s *Session) Set(key, value any) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.session.Values[key] = value
}

func (s *Session) Get(key string) (value any) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return s.session.Values[key]
}

func (s *Session) LockAndUpdate(f func(values map[any]any) error) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	return f(s.session.Values)
}
