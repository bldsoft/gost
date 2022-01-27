package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/bldsoft/gost/config/feature"
	gost_feature "github.com/bldsoft/gost/config/feature"
	"github.com/bldsoft/gost/entity"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/service"
)

//FeatureMongoRepository implements IFeatureRepository interface
type FeatureMongoRepository struct {
	rep         *Repository
	serviceName string
}

// NewFeatureRepository creates feature repository.
func NewFeatureRepository(db *MongoDb, serviceName string) *FeatureMongoRepository {
	rep := &FeatureMongoRepository{rep: NewRepository(db, "feature"), serviceName: serviceName}
	db.AddOnConnectHandler(func() {
		rep.Load()
		log.Infof("Features loaded")
		rep.InitWatcher()
	})
	return rep
}

func (r *FeatureMongoRepository) InitWatcher() {
	w := NewWatcher(r.rep.Collection())
	w.SetHandler(func(fullDocument bson.Raw, optype OperationType) {
		f := &entity.Feature{}
		err := bson.Unmarshal(fullDocument, f)
		if err != nil {
			log.Errorf("Failed to unmarshal Feature: %s", err.Error())
			return
		}
		r.SetFeature(f)
	})
	w.Start()
}

// SetFeature ...
func (r *FeatureMongoRepository) SetFeature(feature *entity.Feature) {
	value := *feature.GlobalValue
	if feature.SrvValues != nil {
		for _, serviceValue := range *feature.SrvValues {
			if serviceValue.SrvName == r.serviceName {
				value = serviceValue.Value
				break
			}
		}
	}

	if f := gost_feature.Features.Get(feature.ID); f != nil {
		f.SetFromString(value)
	}
}

// Load loads features
func (r *FeatureMongoRepository) Load() {
	features := r.GetAll(context.Background())
	for _, feature := range features {
		r.SetFeature(feature)
	}
}

func (r *FeatureMongoRepository) FindByName(ctx context.Context, name string) *entity.Feature {
	item := &entity.Feature{}
	err := r.rep.FindOne(ctx, bson.M{"name": name}, item)
	if err == nil {
		return item
	}
	return nil
}

func (r *FeatureMongoRepository) FindByID(ctx context.Context, id feature.IdType) *entity.Feature {
	item := &entity.Feature{}
	err := r.rep.FindOne(ctx, bson.M{"_id": id}, item)
	if err != nil {
		return nil
	}
	return item
}

func (r *FeatureMongoRepository) GetAll(ctx context.Context) []*entity.Feature {
	var results []*entity.Feature
	r.rep.GetAll(ctx, &results)
	return results
}

func (r *FeatureMongoRepository) Update(ctx context.Context, feature *entity.Feature) error {
	return r.rep.UpdateAndGetByID(ctx, feature, feature)
}

// Compile time checks to ensure your type satisfies an interface
var _ service.IFeatureRepository = (*FeatureMongoRepository)(nil)
