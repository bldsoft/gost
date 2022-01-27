package service

import (
	"context"

	"github.com/bldsoft/gost/config/feature"
	"github.com/bldsoft/gost/entity"
)

// IFeatureRepository ...
type IFeatureRepository interface {
	GetAll(context.Context) []*entity.Feature
	FindByID(ctx context.Context, id feature.IdType) *entity.Feature
	Update(ctx context.Context, features *entity.Feature) error
}

// IFeatureService ...
type IFeatureService interface {
	GetAll(ctx context.Context) []*entity.Feature
	Get(ctx context.Context, id feature.IdType) *entity.Feature
	Update(ctx context.Context, feature *entity.Feature) error
}
