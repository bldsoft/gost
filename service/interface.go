package service

import (
	"context"

	"github.com/bldsoft/gost/config/feature"
	"github.com/bldsoft/gost/entity"
)

type IFeatureService interface {
	GetAll(ctx context.Context) []*entity.Feature
	Get(ctx context.Context, id feature.IdType) *entity.Feature
	Update(ctx context.Context, feature *entity.Feature) error
}
