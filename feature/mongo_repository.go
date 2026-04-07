package feature

import (
	"context"

	config "github.com/bldsoft/gost/config/feature"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/repository"
)

const DefaultCollectionName = "feature"

type mongoFeatureRepository struct {
	rep                 repository.Repository[Feature, *Feature]
	serviceInstanceName string
}

func newMongoFeatureRepository(rep repository.Repository[Feature, *Feature], serviceInstanceName string) *mongoFeatureRepository {
	r := &mongoFeatureRepository{rep: rep, serviceInstanceName: serviceInstanceName}
	if err := r.Load(); err != nil {
		log.Error("Failed to load features")
	} else {
		log.Infof("Features loaded")
	}
	return r
}

func (r *mongoFeatureRepository) SetFeature(feature *Feature) {
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

func (r *mongoFeatureRepository) Load() error {
	features, err := r.GetAll(context.Background())
	if err != nil {
		return err
	}
	for _, feature := range features {
		r.SetFeature(feature)
	}
	return nil
}

func (r *mongoFeatureRepository) FindByName(ctx context.Context, name string) *Feature {
	feature, err := r.rep.FindOne(ctx, map[string]any{"name": name})
	if err != nil {
		return nil
	}
	return feature
}

func (r *mongoFeatureRepository) FindByID(ctx context.Context, id config.IdType) (*Feature, error) {
	return r.rep.FindOne(ctx, map[string]any{"_id": id})
}

func (r *mongoFeatureRepository) GetAll(ctx context.Context) ([]*Feature, error) {
	return r.rep.GetAll(ctx, &repository.QueryOptions{Archived: false})
}

func (r *mongoFeatureRepository) Update(ctx context.Context, feature *Feature) (*Feature, error) {
	return r.rep.UpdateAndGetByID(ctx, feature, true)
}

var _ IFeatureRepository = (*mongoFeatureRepository)(nil)
