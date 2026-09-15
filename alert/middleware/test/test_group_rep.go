package test

import (
	"context"
	"slices"
	"sync"

	"github.com/bldsoft/gost/alert/middleware"
)

type testGroupRepository struct {
	mu     sync.Mutex
	groups map[string]*middleware.Group
}

func newTestGroupRepository() *testGroupRepository {
	return &testGroupRepository{
		groups: make(map[string]*middleware.Group),
	}
}

func (r *testGroupRepository) CreateGroup(ctx context.Context, group *middleware.Group) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.groups[group.ID] = group

	return nil
}

func (r *testGroupRepository) UpdateGroup(ctx context.Context, group *middleware.Group) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.groups[group.ID] = group

	return nil
}

func (r *testGroupRepository) FindGroups(ctx context.Context, filter middleware.GroupFilter) ([]*middleware.Group, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.findGroups(filter), nil
}

func (r *testGroupRepository) Delete(ctx context.Context, filter middleware.GroupFilter) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, group := range r.findGroups(filter) {
		delete(r.groups, group.ID)
	}

	return nil
}

func (r *testGroupRepository) findGroups(filter middleware.GroupFilter) []*middleware.Group {
	groups := make([]*middleware.Group, 0, len(r.groups))
	for _, group := range r.groups {
		if len(filter.IDs) > 0 && !slices.Contains(filter.IDs, group.ID) {
			continue
		}
		if !filter.ExpNotAfter.IsZero() && group.ExpAt.After(filter.ExpNotAfter) {
			continue
		}
		groups = append(groups, group)
	}

	return groups
}
