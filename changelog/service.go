package changelog

import "context"

type Service struct {
	rep IChangeLogRepository
}

func NewService(rep IChangeLogRepository) *Service {
	return &Service{rep}
}

func (s *Service) GetRecords(ctx context.Context, filter *Filter) ([]*Record, error) {
	return s.rep.GetRecords(ctx, filter)
}
