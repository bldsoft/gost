//go:build integration_test

package repository_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	mongodrv "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	gost "github.com/bldsoft/gost/config"
	"github.com/bldsoft/gost/examples/notes/pkg/entity"
	"github.com/bldsoft/gost/examples/notes/pkg/repository"
	"github.com/bldsoft/gost/mongo"
	gostrepo "github.com/bldsoft/gost/repository"
)

var db *mongo.Storage

func TestMain(m *testing.M) {
	var cfg mongo.Config
	gost.ReadConfig(&cfg, "")

	if reachable(cfg) {
		db = mongo.NewStorage(cfg)
		db.Connect()
	}

	code := m.Run()

	if db != nil {
		ctx := context.Background()
		_ = db.Db.Collection(repository.CollectionName).Drop(ctx)
		_ = db.Disconnect(ctx)
	}
	os.Exit(code)
}

func reachable(cfg mongo.Config) bool {
	const timeout = 2 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client, err := mongodrv.Connect(options.Client().ApplyURI(cfg.Server.String()).SetServerSelectionTimeout(timeout))
	if err != nil {
		return false
	}
	defer func() { _ = client.Disconnect(ctx) }()

	return client.Ping(ctx, nil) == nil
}

// Locally a missing database skips the test; in CI (GitHub sets CI=true) it must fail,
// otherwise a broken CI service goes unnoticed.
func requireDB(t *testing.T) *repository.NoteRepository {
	t.Helper()
	if db == nil {
		if os.Getenv("CI") != "" {
			t.Fatal("MongoDB is required in CI: check DATABASE_SERVER")
		}
		t.Skip("MongoDB is not reachable: set DATABASE_SERVER to run integration tests")
	}
	t.Cleanup(func() {
		_, _ = db.Db.Collection(repository.CollectionName).DeleteMany(context.Background(), bson.M{})
	})

	return repository.NewNoteRepository(db)
}

func TestNoteRepository_SoftDeleteAndPurge(t *testing.T) {
	rep := requireDB(t)
	ctx := context.Background()

	note := &entity.Note{Title: "first"}
	require.NoError(t, rep.Insert(ctx, note))
	require.False(t, note.IsZeroID())
	require.NotNil(t, note.CreateTime)

	found, err := rep.FindByID(ctx, note.StringID())
	require.NoError(t, err)
	require.Equal(t, "first", found.Title)

	require.NoError(t, rep.Delete(ctx, note.StringID()))

	_, err = rep.FindByID(ctx, note.StringID())
	require.ErrorIs(t, err, gostrepo.ErrNotFound)

	archived, err := rep.FindByID(ctx, note.StringID(), &gostrepo.QueryOptions{Archived: true})
	require.NoError(t, err)
	require.True(t, archived.IsArchived())

	require.NoError(t, rep.PurgeDeletedBefore(ctx, time.Now().Add(-time.Hour)))
	_, err = rep.FindByID(ctx, note.StringID(), &gostrepo.QueryOptions{Archived: true})
	require.NoError(t, err, "a recently deleted note must survive the purge")

	require.NoError(t, rep.PurgeDeletedBefore(ctx, time.Now().Add(time.Second)))
	_, err = rep.FindByID(ctx, note.StringID(), &gostrepo.QueryOptions{Archived: true})
	require.ErrorIs(t, err, gostrepo.ErrNotFound)
}

func TestNoteRepository_UpdateUnknown(t *testing.T) {
	rep := requireDB(t)

	note := &entity.Note{Title: "ghost"}
	note.GenerateID()
	err := rep.Update(context.Background(), note)
	require.ErrorIs(t, err, gostrepo.ErrNotFound)
}

func TestNoteRepository_ListHidesArchived(t *testing.T) {
	rep := requireDB(t)
	ctx := context.Background()

	kept := &entity.Note{Title: "kept"}
	gone := &entity.Note{Title: "gone"}
	require.NoError(t, rep.Insert(ctx, kept))
	require.NoError(t, rep.Insert(ctx, gone))
	require.NoError(t, rep.Delete(ctx, gone.StringID()))

	visible, err := rep.GetAll(ctx)
	require.NoError(t, err)
	require.Len(t, visible, 1)
	require.Equal(t, "kept", visible[0].Title)

	all, err := rep.GetAll(ctx, &gostrepo.QueryOptions{Archived: true})
	require.NoError(t, err)
	require.Len(t, all, 2)
}
