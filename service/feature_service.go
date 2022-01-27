package service

import (
	"context"

	"github.com/bldsoft/gost/config/feature"
	"github.com/bldsoft/gost/entity"
	"github.com/bldsoft/gost/log"
)

type FeatureService struct {
	featureRep IFeatureRepository
}

func NewFeatureService(featureRep IFeatureRepository) *FeatureService {
	return &FeatureService{featureRep: featureRep}
}

func (srv *FeatureService) Update(ctx context.Context, feature *entity.Feature) error {
	err := srv.validate(feature)
	if err != nil {
		return err
	}

	return srv.featureRep.Update(ctx, feature)
}

func (srv *FeatureService) Get(ctx context.Context, id feature.IdType) *entity.Feature {
	return srv.featureRep.FindByID(ctx, id)
}

func (srv *FeatureService) GetAll(ctx context.Context) []*entity.Feature {
	return srv.featureRep.GetAll(ctx)
}

func (srv *FeatureService) validate(f *entity.Feature) error {
	feature := feature.Features.Get(f.ID)
	if feature == nil {
		log.DebugWithFields(log.Fields{"feature": f.ID}, "Validation of not supported feature")
		return nil
	}

	if f.SrvValues != nil {
		for _, serviceValue := range *f.SrvValues {
			err := feature.Validate(serviceValue.Value)
			if err != nil {
				return err
			}
		}
	}
	if f.GlobalValue != nil {
		return feature.Validate(*f.GlobalValue)
	}
	return nil
}
