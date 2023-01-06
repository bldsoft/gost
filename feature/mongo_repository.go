package feature

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	config "github.com/bldsoft/gost/config/feature"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/mongo"
	"github.com/bldsoft/gost/repository"
)

// MongoRepository implements IFeatureRepository interface
type MongoRepository struct {
	rep         *mongo.Repository[Feature, *Feature, FeatureFilter]
	serviceName string
}

// NewMongoRepository creates feature repository.
func NewMongoRepository(db *mongo.Storage, serviceName string) *MongoRepository {
	rep := &MongoRepository{rep: mongo.NewRepository[Feature, *Feature, FeatureFilter](db, "feature"), serviceName: serviceName}
	if err := rep.Load(); err != nil {
		log.Error("Failed to load features")
	} else {
		log.Infof("Features loaded")
	}
	rep.InitWatcher()
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

	if f := config.Features.Get(feature.ID); f != nil {
		f.SetFromString(value)
	}
}

// Load loads features
func (r *MongoRepository) Load() error {
	features, err := r.GetAll(context.Background())
	if err != nil {
		return err
	}
	for _, feature := range features {
		r.SetFeature(feature)
	}
	return nil
}

func (r *MongoRepository) FindByName(ctx context.Context, name string) *Feature {
	feature, err := r.rep.FindOne(ctx, bson.M{"name": name})
	if err != nil {
		return nil
	}
	return feature
}

func (r *MongoRepository) FindByID(ctx context.Context, id config.IdType) (*Feature, error) {
	feature, err := r.rep.FindOne(ctx, bson.M{"_id": id})
	if err != nil {
		return nil, err
	}
	return feature, nil
}

func (r *MongoRepository) GetAll(ctx context.Context) ([]*Feature, error) {
	return r.rep.GetAll(ctx, &repository.QueryOptions[FeatureFilter]{Archived: false})
}

func (r *MongoRepository) Update(ctx context.Context, feature *Feature) (*Feature, error) {
	return r.rep.UpdateAndGetByID(ctx, feature, true)
}

// Compile time checks to ensure your type satisfies an interface
var _ IFeatureRepository = (*MongoRepository)(nil)
