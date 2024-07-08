package mongo

import (
	"context"

	"github.com/bldsoft/gost/repository"
	"go.mongodb.org/mongo-driver/mongo"
)

//go:generate go run github.com/vektra/mockery/v2 --all --with-expecter

type Repository[T any, U repository.IEntityIDPtr[T]] interface {
	Name() string
	Collection() *mongo.Collection
	WithTransaction(ctx context.Context, f func(ctx mongo.SessionContext) (interface{}, error)) (interface{}, error)

	repository.Repository[T, U]

	AggregateOne(ctx context.Context, pipeline mongo.Pipeline, entity interface{}) error
}
