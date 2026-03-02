package v2

import (
	"time"

	"github.com/bldsoft/gost/repository"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	defaultOpt = repository.QueryOptions{
		Archived: false,
	}
)

var UserEntryCtxKey interface{} = "UserEntry"

type IEntityTimeStamp interface {
	SetUpdateFields(cupdateTime time.Time, updateUserID interface{})
	SetCreateFields(createTime time.Time, createUserID interface{})
}

type BaseRepository[T any, U repository.IEntityIDPtr[T]] struct {
	dbcollection *mongo.Collection
	db
}
