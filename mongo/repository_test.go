//go:build integration_test

package mongo

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/bldsoft/gost/config"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/repository"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

const testCollection = "test_collection"

var (
	db *Storage
)

type testEntity struct {
	EntityID `bson:",inline"`
	Field    string
}

func TestMain(m *testing.M) {
	log.SetLogLevel("")

	var cfg Config
	config.ReadConfig(&cfg, "")

	db = NewStorage(cfg)
	db.Connect()

	code := m.Run()

	ctx := context.Background()
	_, _ = db.Db.Collection(testCollection).DeleteMany(ctx, bson.M{})
	_ = db.Disconnect(ctx)

	os.Exit(code)
}

func testRepository(t *testing.T) (context.Context, Repository[testEntity, *testEntity]) {
	t.Helper()
	ctx := context.Background()
	rep := NewRepository[testEntity](db, testCollection)
	t.Cleanup(func() {
		clearTestCollection(t, ctx)
	})
	return ctx, rep
}

func clearTestCollection(t *testing.T, ctx context.Context) {
	t.Helper()
	_, err := db.Db.Collection(testCollection).DeleteMany(ctx, bson.M{})
	assert.NoError(t, err)
}

func TestBaseRepository_InsertOrReplace(t *testing.T) {
	tests := []struct {
		name         string
		prepare      func(t *testing.T, ctx context.Context, rep Repository[testEntity, *testEntity]) *testEntity
		wantInserted bool
	}{
		{
			name: "insert new entity without id",
			prepare: func(t *testing.T, ctx context.Context, rep Repository[testEntity, *testEntity]) *testEntity {
				return &testEntity{Field: "value1"}
			},
			wantInserted: true,
		},
		{
			name: "insert entity with generated id when document does not exist",
			prepare: func(t *testing.T, ctx context.Context, rep Repository[testEntity, *testEntity]) *testEntity {
				e := &testEntity{Field: "value2"}
				e.GenerateID()
				return e
			},
			wantInserted: true,
		},
		{
			name: "replace existing entity",
			prepare: func(t *testing.T, ctx context.Context, rep Repository[testEntity, *testEntity]) *testEntity {
				e := &testEntity{Field: "initial"}
				assert.NoError(t, rep.Insert(ctx, e))
				e.Field = "updated"
				return e
			},
			wantInserted: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, rep := testRepository(t)
			entity := tt.prepare(t, ctx, rep)

			inserted, err := rep.InsertOrReplace(ctx, entity)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantInserted, inserted)
		})
	}
}

func TestBaseRepository_InsertOrReplaceMany(t *testing.T) {
	t.Run("empty slice does nothing", func(t *testing.T) {
		ctx, rep := testRepository(t)

		err := rep.InsertOrReplaceMany(ctx, []*testEntity{})
		assert.NoError(t, err)

		count, err := rep.Collection().CountDocuments(ctx, bson.M{})
		assert.NoError(t, err)
		assert.EqualValues(t, 0, count)
	})

	t.Run("single new entity uses InsertOrReplace path", func(t *testing.T) {
		ctx, rep := testRepository(t)

		entity := &testEntity{Field: "single_new"}

		err := rep.InsertOrReplaceMany(ctx, []*testEntity{entity})
		assert.NoError(t, err)

		got, err := rep.Find(ctx, bson.M{"field": "single_new"})
		assert.NoError(t, err)
		assert.Len(t, got, 1)
	})

	t.Run("single existing entity is replaced via InsertOrReplace path", func(t *testing.T) {
		ctx, rep := testRepository(t)

		entity := &testEntity{Field: "single_initial"}
		assert.NoError(t, rep.Insert(ctx, entity))

		entity.Field = "single_updated"
		err := rep.InsertOrReplaceMany(ctx, []*testEntity{entity})
		assert.NoError(t, err)

		got, err := rep.FindByID(ctx, entity.StringID())
		assert.NoError(t, err)
		assert.Equal(t, "single_updated", got.Field)

		count, err := rep.Collection().CountDocuments(ctx, bson.M{})
		assert.NoError(t, err)
		assert.EqualValues(t, 1, count)
	})

	t.Run("only new entities without id are inserted", func(t *testing.T) {
		ctx, rep := testRepository(t)

		entities := []*testEntity{
			{Field: "new_a"},
			{Field: "new_b"},
			{Field: "new_c"},
		}

		err := rep.InsertOrReplaceMany(ctx, entities)
		assert.NoError(t, err)

		count, err := rep.Collection().CountDocuments(ctx, bson.M{})
		assert.NoError(t, err)
		assert.EqualValues(t, 3, count)
	})

	t.Run("only existing entities are replaced", func(t *testing.T) {
		ctx, rep := testRepository(t)

		entities := []*testEntity{
			{Field: "ex_a"},
			{Field: "ex_b"},
		}
		for _, e := range entities {
			assert.NoError(t, rep.Insert(ctx, e))
		}

		for _, e := range entities {
			e.Field = e.Field + "_replaced"
		}

		err := rep.InsertOrReplaceMany(ctx, entities)
		assert.NoError(t, err)

		for _, e := range entities {
			got, err := rep.FindByID(ctx, e.StringID())
			assert.NoError(t, err)
			assert.Equal(t, e.Field, got.Field)
		}

		count, err := rep.Collection().CountDocuments(ctx, bson.M{})
		assert.NoError(t, err)
		assert.EqualValues(t, 2, count)
	})

	t.Run("insert and replace multiple entities", func(t *testing.T) {
		ctx, rep := testRepository(t)

		existing1 := &testEntity{Field: "e1"}
		existing2 := &testEntity{Field: "e2"}
		assert.NoError(t, rep.Insert(ctx, existing1))
		assert.NoError(t, rep.Insert(ctx, existing2))

		newEntity := &testEntity{Field: "new"}
		existing1.Field = "e1_updated"
		existing2.Field = "e2_updated"

		err := rep.InsertOrReplaceMany(ctx, []*testEntity{
			newEntity,
			existing1,
			existing2,
		})
		assert.NoError(t, err)

		got1, err := rep.FindByID(ctx, existing1.StringID())
		assert.NoError(t, err)
		assert.Equal(t, "e1_updated", got1.Field)

		got2, err := rep.FindByID(ctx, existing2.StringID())
		assert.NoError(t, err)
		assert.Equal(t, "e2_updated", got2.Field)

		foundNew, err := rep.Find(ctx, bson.M{"field": "new"})
		assert.NoError(t, err)
		assert.Len(t, foundNew, 1)
	})
}

func TestBaseRepository_Replace(t *testing.T) {
	t.Run("replaces existing entity", func(t *testing.T) {
		ctx, rep := testRepository(t)

		entity := &testEntity{Field: "initial"}
		assert.NoError(t, rep.Insert(ctx, entity))

		entity.Field = "replaced"
		err := rep.Replace(ctx, entity)
		assert.NoError(t, err)

		got, err := rep.FindByID(ctx, entity.StringID())
		assert.NoError(t, err)
		assert.Equal(t, "replaced", got.Field)
	})

	t.Run("returns ErrNotFound when entity does not exist", func(t *testing.T) {
		ctx, rep := testRepository(t)

		entity := &testEntity{Field: "missing"}
		entity.GenerateID()

		err := rep.Replace(ctx, entity)
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("returns ErrNotFound when id is zero and collection is empty", func(t *testing.T) {
		ctx, rep := testRepository(t)

		entity := &testEntity{Field: "zero_id"}

		err := rep.Replace(ctx, entity)
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})
}

func TestBaseRepository_ReplaceOne(t *testing.T) {
	t.Run("replaces by custom filter", func(t *testing.T) {
		ctx, rep := testRepository(t)

		entity := &testEntity{Field: "by_field"}
		assert.NoError(t, rep.Insert(ctx, entity))

		replacement := &testEntity{EntityID: entity.EntityID, Field: "by_field_replaced"}
		err := rep.ReplaceOne(ctx, bson.M{"field": "by_field"}, replacement)
		assert.NoError(t, err)

		got, err := rep.FindByID(ctx, entity.StringID())
		assert.NoError(t, err)
		assert.Equal(t, "by_field_replaced", got.Field)
	})

	t.Run("returns ErrNotFound when filter matches nothing", func(t *testing.T) {
		ctx, rep := testRepository(t)

		entity := &testEntity{Field: "no_match"}
		entity.GenerateID()

		err := rep.ReplaceOne(ctx, bson.M{"field": "does_not_exist"}, entity)
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})
}

func TestBaseRepository_InsertOrReplaceOne(t *testing.T) {
	t.Run("inserts when filter matches nothing (upsert)", func(t *testing.T) {
		ctx, rep := testRepository(t)

		entity := &testEntity{Field: "upserted"}
		entity.GenerateID()

		inserted, err := rep.InsertOrReplaceOne(ctx, bson.M{"_id": entity.RawID()}, entity)
		assert.NoError(t, err)
		assert.True(t, inserted)

		got, err := rep.FindByID(ctx, entity.StringID())
		assert.NoError(t, err)
		assert.Equal(t, "upserted", got.Field)
	})

	t.Run("replaces when filter matches existing document", func(t *testing.T) {
		ctx, rep := testRepository(t)

		entity := &testEntity{Field: "to_be_replaced"}
		assert.NoError(t, rep.Insert(ctx, entity))

		replacement := &testEntity{EntityID: entity.EntityID, Field: "after_replace"}
		inserted, err := rep.InsertOrReplaceOne(ctx, bson.M{"_id": entity.RawID()}, replacement)
		assert.NoError(t, err)
		assert.False(t, inserted)

		got, err := rep.FindByID(ctx, entity.StringID())
		assert.NoError(t, err)
		assert.Equal(t, "after_replace", got.Field)

		count, err := rep.Collection().CountDocuments(ctx, bson.M{})
		assert.NoError(t, err)
		assert.EqualValues(t, 1, count)
	})

	t.Run("upserts by custom filter", func(t *testing.T) {
		ctx, rep := testRepository(t)

		entity := &testEntity{Field: "custom_filter"}
		entity.GenerateID()

		inserted, err := rep.InsertOrReplaceOne(ctx, bson.M{"field": "custom_filter"}, entity)
		assert.NoError(t, err)
		assert.True(t, inserted)

		got, err := rep.Find(ctx, bson.M{"field": "custom_filter"})
		assert.NoError(t, err)
		assert.Len(t, got, 1)
	})
}

func TestBaseRepository_InsertOrReplace_ReplaceWithExplicitID(t *testing.T) {
	ctx, rep := testRepository(t)

	entity := &testEntity{Field: "explicit_initial"}
	entity.GenerateID()
	originalID := entity.ID

	assert.NoError(t, rep.Insert(ctx, entity))

	entity.Field = "explicit_updated"
	inserted, err := rep.InsertOrReplace(ctx, entity)
	assert.NoError(t, err)
	assert.False(t, inserted)
	assert.Equal(t, originalID, entity.ID)

	got, err := rep.FindByID(ctx, entity.StringID())
	assert.NoError(t, err)
	assert.Equal(t, "explicit_updated", got.Field)

	count, err := rep.Collection().CountDocuments(ctx, bson.M{})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, count)
}

func TestBaseRepository_InsertOrReplace_InsertWithIDForUnknownDocument(t *testing.T) {
	ctx, rep := testRepository(t)

	entity := &testEntity{Field: "preset_id"}
	entity.GenerateID()

	inserted, err := rep.InsertOrReplace(ctx, entity)
	assert.NoError(t, err)
	assert.True(t, inserted)

	got, err := rep.FindByID(ctx, entity.StringID())
	assert.NoError(t, err)
	assert.Equal(t, "preset_id", got.Field)
}

func TestBaseRepository_InsertOrReplace_DoesNotReturnErrNotFound(t *testing.T) {
	ctx, rep := testRepository(t)

	entity := &testEntity{Field: "any"}
	entity.GenerateID()

	_, err := rep.InsertOrReplace(ctx, entity)
	assert.NoError(t, err)
	assert.False(t, errors.Is(err, repository.ErrNotFound))
}
