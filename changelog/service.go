package changelog

import (
	"context"
)

type Service struct {
	rep IChangeLogRepository
}

func NewService(rep IChangeLogRepository) *Service {
	return &Service{rep}
}

func (s *Service) FindByID(ctx context.Context, id string) (*Record, error) {
	return s.rep.FindByID(ctx, id)
}

func (s *Service) FindByIDs(ctx context.Context, ids []string, preserveOrder bool) (res []*Record, err error) {
	return s.rep.FindByIDs(ctx, ids, preserveOrder)
}

func (s *Service) GetRecords(ctx context.Context, filter *Filter) ([]*Record, error) {
	return s.rep.GetRecords(ctx, filter)
}
