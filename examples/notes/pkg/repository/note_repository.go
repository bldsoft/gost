package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/bldsoft/gost/examples/notes/pkg/entity"
	"github.com/bldsoft/gost/mongo"
	"github.com/bldsoft/gost/repository"
)

const CollectionName = "notes"

type NoteRepository struct {
	mongo.Repository[entity.Note, *entity.Note]
}

func NewNoteRepository(db *mongo.Storage) *NoteRepository {
	return &NoteRepository{Repository: mongo.NewRepository[entity.Note](db, CollectionName)}
}

func (r *NoteRepository) PurgeDeletedBefore(ctx context.Context, before time.Time) error {
	filter := bson.M{mongo.BsonFieldNameDeleteTime: bson.M{"$lt": before}}

	// Without an explicit Archived:false option DeleteMany only marks documents as archived.
	return r.DeleteMany(ctx, filter, &repository.QueryOptions{Archived: false})
}
