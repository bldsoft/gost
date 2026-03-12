package feature

import (
	"context"

	config "github.com/bldsoft/gost/config/feature"
	"github.com/bldsoft/gost/log"
	mongov2 "github.com/bldsoft/gost/mongo/v2"
	"github.com/bldsoft/gost/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MongoRepositoryV2 struct {
	rep                 mongov2.Repository[Feature, *Feature]
	serviceInstanceName string
}

func NewMongoRepositoryV2(db *mongov2.Storage, serviceInstanceName string, collName ...string) *MongoRepositoryV2 {
	if len(collName) == 0 {
		collName = []string{DefaultCollectionName}
	}
	rep := &MongoRepositoryV2{rep: mongov2.NewRepository[Feature](db, collName[0]), serviceInstanceName: serviceInstanceName}
	if err := rep.Load(); err != nil {
		log.Error("Failed to load features")
	} else {
		log.Infof("Features loaded")
	}
	rep.InitWatcher()
	return rep
}

func (r *MongoRepositoryV2) InitWatcher() {
	w := mongov2.NewWatcher(r.rep.Collection())
	w.SetHandler(func(fullDocument bson.Raw, optype mongov2.OperationType) {
		f := &Feature{}
		err := bson.Unmarshal(fullDocument, f)
		if err != nil {
			log.Errorf("Failed to unmarshal Feature: %s", err.Error())
			return
		}
		r.SetFeature(f)
	})
	w.Start()
}

func (r *MongoRepositoryV2) SetFeature(feature *Feature) {
	value := *feature.GlobalValue
	if feature.SrvValues != nil {
		for _, serviceValue := range *feature.SrvValues {
			if serviceValue.SrvName == r.serviceInstanceName {
				value = serviceValue.Value
				break
			}
		}
	}

	if f := config.Features.Get(feature.ID); f != nil {
		f.SetFromString(value)
	}
}

func (r *MongoRepositoryV2) Load() error {
	features, err := r.GetAll(context.Background())
	if err != nil {
		return err
	}
	for _, feature := range features {
		r.SetFeature(feature)
	}
	return nil
}

func (r *MongoRepositoryV2) FindByName(ctx context.Context, name string) *Feature {
	feature, err := r.rep.FindOne(ctx, bson.M{"name": name})
	if err != nil {
		return nil
	}
	return feature
}

func (r *MongoRepositoryV2) FindByID(ctx context.Context, id config.IdType) (*Feature, error) {
	feature, err := r.rep.FindOne(ctx, bson.M{"_id": id})
	if err != nil {
		return nil, err
	}
	return feature, nil
}

func (r *MongoRepositoryV2) GetAll(ctx context.Context) ([]*Feature, error) {
	return r.rep.GetAll(ctx, &repository.QueryOptions{Archived: false})
}

func (r *MongoRepositoryV2) Update(ctx context.Context, feature *Feature) (*Feature, error) {
	return r.rep.UpdateAndGetByID(ctx, feature, true)
}

var _ IFeatureRepository = (*MongoRepositoryV2)(nil)
