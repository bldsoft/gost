package feature

import (
	"context"

	"github.com/bldsoft/gost/config/feature"
	"github.com/bldsoft/gost/log"
)

type Service struct {
	featureRep IFeatureRepository
}

func NewService(featureRep IFeatureRepository) *Service {
	return &Service{featureRep: featureRep}
}

func (srv *Service) Update(ctx context.Context, feature *Feature) error {
	err := srv.validate(feature)
	if err != nil {
		return err
	}

	return srv.featureRep.Update(ctx, feature)
}

func (srv *Service) Get(ctx context.Context, id feature.IdType) *Feature {
	return srv.featureRep.FindByID(ctx, id)
}

func (srv *Service) GetAll(ctx context.Context) []*Feature {
	return srv.featureRep.GetAll(ctx)
}

func (srv *Service) validate(f *Feature) error {
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
