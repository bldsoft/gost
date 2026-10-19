package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/bldsoft/gost/repository"
)

//go:generate mockery

type Repository[T any, U repository.IEntityIDPtr[T]] interface {
	Name() string
	Collection() *mongo.Collection
	WithTransaction(ctx context.Context, f func(ctx context.Context) (any, error)) (any, error)

	repository.Repository[T, U]

	AggregateOne(ctx context.Context, pipeline mongo.Pipeline, entity any) error
}
