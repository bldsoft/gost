package feature

import (
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func NewMongoRepository(db *mongo.Storage, serviceInstanceName string, collName ...string) *mongoFeatureRepository {
	if len(collName) == 0 {
		collName = []string{DefaultCollectionName}
	}
	rep := mongo.NewRepository[Feature](db, collName[0])
	r := newMongoFeatureRepository(rep, serviceInstanceName)

	w := mongo.NewWatcher(rep.Collection())
	w.SetHandler(func(fullDocument bson.Raw, optype mongo.OperationType) {
		f := &Feature{}
		if err := bson.Unmarshal(fullDocument, f); err != nil {
			log.Errorf("Failed to unmarshal Feature: %s", err.Error())
			return
		}
		r.SetFeature(f)
	})
	w.Start()

	return r
}
