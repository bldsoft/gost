package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/bldsoft/gost/examples/notes/pkg/entity"
	"github.com/bldsoft/gost/examples/notes/pkg/service"
	"github.com/bldsoft/gost/examples/notes/pkg/service/mocks"
	"github.com/bldsoft/gost/repository"
)

func TestNoteService_Create(t *testing.T) {
	t.Run("trims the title and inserts", func(t *testing.T) {
		repo := mocks.NewNoteRepository(t)
		note := &entity.Note{Title: "  hello  "}
		repo.EXPECT().Insert(mock.Anything, note).Return(nil)

		require.NoError(t, service.NewNoteService(repo).Create(context.Background(), note))
		require.Equal(t, "hello", note.Title)
	})

	t.Run("rejects an empty title without touching the repository", func(t *testing.T) {
		repo := mocks.NewNoteRepository(t)

		err := service.NewNoteService(repo).Create(context.Background(), &entity.Note{Title: "   "})
		require.ErrorIs(t, err, entity.ErrInvalidNote)
	})

	t.Run("passes repository errors through", func(t *testing.T) {
		repo := mocks.NewNoteRepository(t)
		repo.EXPECT().Insert(mock.Anything, mock.Anything).Return(repository.ErrAlreadyExists)

		err := service.NewNoteService(repo).Create(context.Background(), &entity.Note{Title: "dup"})
		require.ErrorIs(t, err, repository.ErrAlreadyExists)
	})
}

func TestNoteService_Get(t *testing.T) {
	repo := mocks.NewNoteRepository(t)
	id := bson.NewObjectID()
	want := &entity.Note{Title: "found"}
	want.ID = id
	repo.EXPECT().FindByID(mock.Anything, id.Hex(), []*repository.QueryOptions{{Archived: true}}).Return(want, nil)

	got, err := service.NewNoteService(repo).Get(context.Background(), id.Hex(), true)
	require.NoError(t, err)
	require.Same(t, want, got)
}

func TestNoteService_List(t *testing.T) {
	repo := mocks.NewNoteRepository(t)
	repo.EXPECT().GetAll(mock.Anything, mock.MatchedBy(func(opts []*repository.QueryOptions) bool {
		opt := opts[0]

		return len(opts) == 1 && opt.Limit == 10 && !opt.Archived && len(opt.Sort) == 1 && opt.Sort[0].Desc
	})).Return([]*entity.Note{}, nil)

	notes, err := service.NewNoteService(repo).List(context.Background(), 10, false)
	require.NoError(t, err)
	require.Empty(t, notes)
}

func TestNoteService_Update(t *testing.T) {
	t.Run("sets the id from the path and updates", func(t *testing.T) {
		repo := mocks.NewNoteRepository(t)
		id := bson.NewObjectID()
		note := &entity.Note{Title: "new"}
		repo.EXPECT().Update(mock.Anything, note).Return(nil)

		require.NoError(t, service.NewNoteService(repo).Update(context.Background(), id.Hex(), note))
		require.Equal(t, id, note.ID)
	})

	t.Run("rejects a malformed id", func(t *testing.T) {
		repo := mocks.NewNoteRepository(t)

		err := service.NewNoteService(repo).Update(context.Background(), "not-an-object-id", &entity.Note{Title: "x"})
		require.ErrorIs(t, err, entity.ErrInvalidNote)
	})
}

func TestNoteService_Delete(t *testing.T) {
	repo := mocks.NewNoteRepository(t)
	repo.EXPECT().Delete(mock.Anything, "abc").Return(repository.ErrNotFound)

	err := service.NewNoteService(repo).Delete(context.Background(), "abc")
	require.ErrorIs(t, err, repository.ErrNotFound)
}

func TestPurger(t *testing.T) {
	repo := mocks.NewNoteRepository(t)
	purged := make(chan time.Time, 1)
	repo.EXPECT().PurgeDeletedBefore(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, before time.Time) error {
			select {
			case purged <- before:
			default:
			}

			return errors.New("ignored by the purger loop")
		})

	purger := service.NewPurger(repo, 5*time.Millisecond, time.Hour)
	done := make(chan error, 1)
	go func() { done <- purger.Run() }()

	select {
	case before := <-purged:
		require.WithinDuration(t, time.Now().Add(-time.Hour), before, time.Minute)
	case <-time.After(5 * time.Second):
		t.Fatal("purger did not run")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	require.NoError(t, purger.Stop(ctx))
	require.NoError(t, <-done)
}
