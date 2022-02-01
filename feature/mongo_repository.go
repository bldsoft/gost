package feature

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/bldsoft/gost/config/feature"
	gost_feature "github.com/bldsoft/gost/config/feature"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/mongo"
)

//MongoRepository implements IFeatureRepository interface
type MongoRepository struct {
	rep         *mongo.Repository
	serviceName string
}

// NewMongoRepository creates feature repository.
func NewMongoRepository(db *mongo.MongoDb, serviceName string) *MongoRepository {
	rep := &MongoRepository{rep: mongo.NewRepository(db, "feature"), serviceName: serviceName}
	db.AddOnConnectHandler(func() {
		rep.Load()
		log.Infof("Features loaded")
		rep.InitWatcher()
	})
	return rep
}

func (r *MongoRepository) InitWatcher() {
	w := mongo.NewWatcher(r.rep.Collection())
	w.SetHandler(func(fullDocument bson.Raw, optype mongo.OperationType) {
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

// SetFeature ...
func (r *MongoRepository) SetFeature(feature *Feature) {
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
func (r *MongoRepository) Load() {
	features := r.GetAll(context.Background())
	for _, feature := range features {
		r.SetFeature(feature)
	}
}

func (r *MongoRepository) FindByName(ctx context.Context, name string) *Feature {
	item := &Feature{}
	err := r.rep.FindOne(ctx, bson.M{"name": name}, item)
	if err == nil {
		return item
	}
	return nil
}

func (r *MongoRepository) FindByID(ctx context.Context, id feature.IdType) *Feature {
	item := &Feature{}
	err := r.rep.FindOne(ctx, bson.M{"_id": id}, item)
	if err != nil {
		return nil
	}
	return item
}

func (r *MongoRepository) GetAll(ctx context.Context) []*Feature {
	var results []*Feature
	r.rep.GetAll(ctx, &results)
	return results
}

func (r *MongoRepository) Update(ctx context.Context, feature *Feature) error {
	return r.rep.UpdateAndGetByID(ctx, feature, feature)
}

// Compile time checks to ensure your type satisfies an interface
var _ IFeatureRepository = (*MongoRepository)(nil)
