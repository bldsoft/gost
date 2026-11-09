package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/bldsoft/gost/examples/notes/pkg/entity"
	"github.com/bldsoft/gost/repository"
)

const maxTitleLen = 200

type NoteService struct {
	notes NoteRepository
}

func NewNoteService(notes NoteRepository) *NoteService {
	return &NoteService{notes: notes}
}

func (s *NoteService) Create(ctx context.Context, note *entity.Note) error {
	if err := validate(note); err != nil {
		return err
	}

	return s.notes.Insert(ctx, note)
}

func (s *NoteService) Get(ctx context.Context, id string, archived bool) (*entity.Note, error) {
	return s.notes.FindByID(ctx, id, &repository.QueryOptions{Archived: archived})
}

func (s *NoteService) List(ctx context.Context, limit int64, archived bool) ([]*entity.Note, error) {
	return s.notes.GetAll(ctx, &repository.QueryOptions{
		Archived: archived,
		Limit:    limit,
		Sort:     repository.Sort().Desc("createTime"),
	})
}

func (s *NoteService) Update(ctx context.Context, id string, note *entity.Note) error {
	if err := note.SetIDFromString(id); err != nil {
		return fmt.Errorf("%w: id %q: %w", entity.ErrInvalidNote, id, err)
	}
	if err := validate(note); err != nil {
		return err
	}

	return s.notes.Update(ctx, note)
}

func (s *NoteService) Delete(ctx context.Context, id string) error {
	return s.notes.Delete(ctx, id)
}

func validate(note *entity.Note) error {
	note.Title = strings.TrimSpace(note.Title)
	if note.Title == "" {
		return fmt.Errorf("%w: title is required", entity.ErrInvalidNote)
	}
	if len(note.Title) > maxTitleLen {
		return fmt.Errorf("%w: title is longer than %d characters", entity.ErrInvalidNote, maxTitleLen)
	}

	return nil
}
