package changelog

import "context"

type Service struct {
	rep IChangeLogRepository
}

func NewService(rep IChangeLogRepository) *Service {
	return &Service{rep}
}

func (s *Service) GetByID(ctx context.Context, id idType) (*Record, error) {
	return s.rep.GetByID(ctx, id)
}

func (s *Service) GetRecords(ctx context.Context, filter *Filter) ([]*Record, error) {
	return s.rep.GetRecords(ctx, filter)
}

func (s *Service) GetByIDs(ctx context.Context, ids []interface{}) (res []*Record, err error) {
	return s.rep.GetByIDs(ctx, ids)
}
