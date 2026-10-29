package controller

import (
	"context"

	"github.com/bldsoft/gost/examples/notes/pkg/entity"
)

type NoteService interface {
	Create(ctx context.Context, note *entity.Note) error
	Get(ctx context.Context, id string, archived bool) (*entity.Note, error)
	List(ctx context.Context, limit int64, archived bool) ([]*entity.Note, error)
	Update(ctx context.Context, id string, note *entity.Note) error
	Delete(ctx context.Context, id string) error
}
