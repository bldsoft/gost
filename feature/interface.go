package feature

import (
	"context"

	"github.com/bldsoft/gost/config/feature"
)

// IFeatureRepository ...
type IFeatureRepository interface {
	GetAll(context.Context) []*Feature
	FindByID(ctx context.Context, id feature.IdType) *Feature
	Update(ctx context.Context, features *Feature) error
}

// IFeatureService ...
type IFeatureService interface {
	GetAll(ctx context.Context) []*Feature
	Get(ctx context.Context, id feature.IdType) *Feature
	Update(ctx context.Context, feature *Feature) error
}
