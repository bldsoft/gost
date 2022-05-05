package feature

import (
	"context"

	"github.com/bldsoft/gost/config/feature"
)

// IFeatureRepository ...
type IFeatureRepository interface {
	GetAll(context.Context) ([]*Feature, error)
	FindByID(ctx context.Context, id feature.IdType) (*Feature, error)
	Update(ctx context.Context, feature *Feature) (*Feature, error)
}

// IFeatureService ...
type IFeatureService interface {
	GetAll(ctx context.Context) ([]*Feature, error)
	Get(ctx context.Context, id feature.IdType) (*Feature, error)
	Update(ctx context.Context, features *Feature) (*Feature, error)
}
