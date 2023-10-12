package mongo

import (
	"context"

	"github.com/bldsoft/gost/repository"
	"go.mongodb.org/mongo-driver/mongo"
)

// NOTE: I used golang/mock@v1.7.0-rc.1 to generate mocks for mongo
// the probleim is it treats T withing repository.IEntityIDPtr as a custom type from this package
// meaning, it generates the following signature for mock:
// MockRepository[T any, U repository.IEntityIDPtr[mongo.T]] struct
// if regeneration is neaded, make sure to replace the resulting signature with
// MockRepository[T any, U repository.IEntityIDPtr[T]] struct
type Repository[T any, U repository.IEntityIDPtr[T]] interface {
	Name() string
	Collection() *mongo.Collection
	WithTransaction(ctx context.Context, f func(ctx mongo.SessionContext) (interface{}, error)) (interface{}, error)

	repository.Repository[T, U]

	AggregateOne(ctx context.Context, pipeline mongo.Pipeline, entity interface{}) error
}
