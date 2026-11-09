package service

import (
	"context"
	"time"

	"github.com/bldsoft/gost/examples/notes/pkg/entity"
	"github.com/bldsoft/gost/repository"
)

//go:generate go tool mockery

type NoteRepository interface {
	FindByID(ctx context.Context, id any, opts ...*repository.QueryOptions) (*entity.Note, error)
	GetAll(ctx context.Context, opts ...*repository.QueryOptions) ([]*entity.Note, error)
	Insert(ctx context.Context, note *entity.Note) error
	Update(ctx context.Context, note *entity.Note, opts ...*repository.QueryOptions) error
	Delete(ctx context.Context, id any, opts ...*repository.QueryOptions) error
	PurgeDeletedBefore(ctx context.Context, before time.Time) error
}
