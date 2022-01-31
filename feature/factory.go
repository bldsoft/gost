package feature

import (
	"reflect"

	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/mongo"
	"github.com/bldsoft/gost/storage"
)

func CreateFeatureService(s storage.IStorage, serviceName string) IFeatureService {
	switch s := s.(type) {
	case *mongo.MongoDb:
		return NewFeatureService(NewFeatureMongoRepository(s, serviceName))
	default:
		log.Fatalf("%s doesn't support feature repository", reflect.TypeOf(s))
		return nil
	}
}
