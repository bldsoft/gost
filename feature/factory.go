package feature

import (
	"reflect"

	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/mongo"
	mongov2 "github.com/bldsoft/gost/mongo/v2"
	"github.com/bldsoft/gost/storage"
)

func CreateController(s storage.IStorage, serviceName string) *Controller {
	return NewController(CreateService(s, serviceName))
}

func CreateService(s storage.IStorage, serviceName string) IFeatureService {
	switch s := s.(type) {
	case *mongo.Storage:
		return NewService(NewMongoRepository(s, serviceName))
	case *mongov2.Storage:
		return NewService(NewMongoRepositoryV2(s, serviceName))
	default:
		log.Panicf("%s doesn't support feature repository", reflect.TypeOf(s))
		return nil
	}
}
