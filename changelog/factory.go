package changelog

import (
	"reflect"

	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/mongo"
	"github.com/bldsoft/gost/storage"
)

func CreateRepository(s storage.IStorage) IChangeLogRepository {
	switch s := s.(type) {
	case *mongo.MongoDb:
		return NewChangeLogMongoRepository(s)
	default:
		log.Fatalf("%s doesn't support change log repository", reflect.TypeOf(s))
		return nil
	}
}

func CreateController(rep IChangeLogRepository) *Controller {
	return NewController(NewService(rep))
}
