package store

import (
	"context"

	"github.com/gorilla/sessions"
)

type SessionRepository interface {
	sessions.Store
	AllSessions(ctx context.Context, name string, offset, limit int) ([]*sessions.Session, error)
	SessionByIDs(ctx context.Context, name string, ids ...string) ([]*sessions.Session, error)
	KillSessions(ctx context.Context, sessionIDs ...string) error
	KillUserSessions(ctx context.Context, userID string) error
}
