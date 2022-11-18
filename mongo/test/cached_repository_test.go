////go:build integration_test

package test

import (
	"context"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/bldsoft/gost/cache/bigcache"
	"github.com/bldsoft/gost/config"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/mongo"
	"github.com/bldsoft/gost/repository"
	"github.com/stretchr/testify/assert"
)

const (
	testCollection          = "test"
	waitCacheUpdateDuration = 300 * time.Millisecond
)

var (
	db *mongo.Storage
)

type testEntity struct {
	mongo.EntityID `bson:",inline"`
	Field          string
	Field2         string
}

func generateEntities(n int) []*testEntity {
	entities := make([]*testEntity, 0, n)
	for i := 0; i < n; i++ {
		entities = append(entities, &testEntity{
			Field:  strconv.Itoa(rand.Int()),
			Field2: strconv.Itoa(rand.Int()),
		})
	}
	return entities
}

func testEntityRepository() *mongo.Repository[testEntity, *testEntity] {
	return mongo.NewRepository[testEntity](db, testCollection)
}

func cachedTestEntityRepository() *mongo.CachedRepository[testEntity, *testEntity] {
	cache := bigcache.NewRepository("{}")
	return mongo.NewCachedRepository[testEntity](db, testCollection, cache, "prefix")
}

func setup() {
	log.SetLogLevel("")
	var cfg mongo.Config
	config.ReadConfig(&cfg, "")
	db = mongo.NewStorage(cfg)
	db.Connect()
	// wait connection
	waitC := make(chan struct{})
	db.AddOnConnectHandler(func() {
		close(waitC)
	})
	<-waitC
}

func tearDown() {
	ctx := context.Background()
	db.Db.Collection(testCollection).Drop(ctx)
	_ = db.Disconnect(ctx)
}

func TestMain(m *testing.M) {
	setup()
	exitVal := m.Run()
	tearDown()
	os.Exit(exitVal)
}

func TestFindByID(t *testing.T) {
	ctx := context.Background()
	entity := generateEntities(1)[0]
	assert.NoError(t, testEntityRepository().Insert(ctx, entity))
	time.Sleep(waitCacheUpdateDuration)

	actual, err := cachedTestEntityRepository().FindByID(ctx, entity)
	assert.NoError(t, err)
	assert.Equal(t, entity, actual)
}

func TestFindByIDs(t *testing.T) {
	const n = 3
	ctx := context.Background()
	entities := generateEntities(n)
	assert.NoError(t, testEntityRepository().InsertMany(ctx, entities))
	time.Sleep(waitCacheUpdateDuration)

	ids := make([]interface{}, 0, n)
	for _, e := range entities {
		ids = append(ids, &testEntity{EntityID: e.EntityID})
	}
	actual, err := cachedTestEntityRepository().FindByIDs(ctx, ids, true)
	assert.NoError(t, err)
	assert.Equal(t, entities, actual)
}

func TestFindByStringIDs(t *testing.T) {
	const n = 3
	ctx := context.Background()
	entities := generateEntities(n)
	assert.NoError(t, testEntityRepository().InsertMany(ctx, entities))
	time.Sleep(waitCacheUpdateDuration)

	ids := make([]string, 0, n)
	for _, e := range entities {
		ids = append(ids, e.StringID())
	}
	actual, err := cachedTestEntityRepository().FindByStringIDs(ctx, ids, true)
	assert.NoError(t, err)
	assert.Equal(t, entities, actual)
}

func TestEntityUpdate(t *testing.T) {
	ctx := context.Background()
	entity := generateEntities(1)[0]
	rep := testEntityRepository()
	cachedRep := cachedTestEntityRepository()
	assert.NoError(t, rep.Insert(ctx, entity))
	time.Sleep(waitCacheUpdateDuration)

	entity.Field = "123"
	assert.NoError(t, rep.Update(ctx, entity))
	time.Sleep(waitCacheUpdateDuration)

	actual, err := cachedRep.FindByID(ctx, entity)
	assert.NoError(t, err)
	assert.Equal(t, entity, actual)
}

func TestEntityDelete(t *testing.T) {
	ctx := context.Background()
	rep := testEntityRepository()
	cachedRep := cachedTestEntityRepository()

	entity := generateEntities(1)[0]
	assert.NoError(t, rep.Insert(ctx, entity))
	time.Sleep(waitCacheUpdateDuration)

	actual, err := cachedRep.FindByID(ctx, entity)
	assert.NoError(t, err)
	assert.Equal(t, entity, actual)

	assert.NoError(t, rep.Delete(ctx, entity, &repository.QueryOptions{Archived: false}))
	time.Sleep(waitCacheUpdateDuration)

	actual, err = cachedRep.FindByID(ctx, entity)
	assert.Error(t, err)
	assert.Nil(t, actual)
}

func BenchmarkSingleEntity(b *testing.B) {
	n := 10000
	ctx := context.Background()
	rep := testEntityRepository()
	entities := generateEntities(n)
	entity := entities[0]
	assert.NoError(b, rep.InsertMany(ctx, entities))
	cachedRep := cachedTestEntityRepository()

	b.Run("no cache", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = rep.FindByID(ctx, entity.StringID())
		}
	})
	b.Run("cache", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = cachedRep.FindByID(ctx, entity.StringID())
		}
	})
}
