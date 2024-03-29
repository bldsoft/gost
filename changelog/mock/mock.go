// Code generated by MockGen. DO NOT EDIT.
// Source: changelog/interface.go

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	changelog "github.com/bldsoft/gost/changelog"
	repository "github.com/bldsoft/gost/repository"
	gomock "github.com/golang/mock/gomock"
)

// MockIChangeLogRepository is a mock of IChangeLogRepository interface.
type MockIChangeLogRepository struct {
	ctrl     *gomock.Controller
	recorder *MockIChangeLogRepositoryMockRecorder
}

// MockIChangeLogRepositoryMockRecorder is the mock recorder for MockIChangeLogRepository.
type MockIChangeLogRepositoryMockRecorder struct {
	mock *MockIChangeLogRepository
}

// NewMockIChangeLogRepository creates a new mock instance.
func NewMockIChangeLogRepository(ctrl *gomock.Controller) *MockIChangeLogRepository {
	mock := &MockIChangeLogRepository{ctrl: ctrl}
	mock.recorder = &MockIChangeLogRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIChangeLogRepository) EXPECT() *MockIChangeLogRepositoryMockRecorder {
	return m.recorder
}

// FindByID mocks base method.
func (m *MockIChangeLogRepository) FindByID(ctx context.Context, id string, options ...*repository.QueryOptions) (*changelog.Record, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, id}
	for _, a := range options {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "FindByID", varargs...)
	ret0, _ := ret[0].(*changelog.Record)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindByID indicates an expected call of FindByID.
func (mr *MockIChangeLogRepositoryMockRecorder) FindByID(ctx, id interface{}, options ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, id}, options...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindByID", reflect.TypeOf((*MockIChangeLogRepository)(nil).FindByID), varargs...)
}

// FindByIDs mocks base method.
func (m *MockIChangeLogRepository) FindByIDs(ctx context.Context, ids []string, preserveOrder bool, options ...*repository.QueryOptions) ([]*changelog.Record, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, ids, preserveOrder}
	for _, a := range options {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "FindByIDs", varargs...)
	ret0, _ := ret[0].([]*changelog.Record)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindByIDs indicates an expected call of FindByIDs.
func (mr *MockIChangeLogRepositoryMockRecorder) FindByIDs(ctx, ids, preserveOrder interface{}, options ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, ids, preserveOrder}, options...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindByIDs", reflect.TypeOf((*MockIChangeLogRepository)(nil).FindByIDs), varargs...)
}

// GetRecords mocks base method.
func (m *MockIChangeLogRepository) GetRecords(ctx context.Context, params *changelog.RecordsParams) (*changelog.Records, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRecords", ctx, params)
	ret0, _ := ret[0].(*changelog.Records)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRecords indicates an expected call of GetRecords.
func (mr *MockIChangeLogRepositoryMockRecorder) GetRecords(ctx, params interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRecords", reflect.TypeOf((*MockIChangeLogRepository)(nil).GetRecords), ctx, params)
}

// MockIChangeLogService is a mock of IChangeLogService interface.
type MockIChangeLogService struct {
	ctrl     *gomock.Controller
	recorder *MockIChangeLogServiceMockRecorder
}

// MockIChangeLogServiceMockRecorder is the mock recorder for MockIChangeLogService.
type MockIChangeLogServiceMockRecorder struct {
	mock *MockIChangeLogService
}

// NewMockIChangeLogService creates a new mock instance.
func NewMockIChangeLogService(ctrl *gomock.Controller) *MockIChangeLogService {
	mock := &MockIChangeLogService{ctrl: ctrl}
	mock.recorder = &MockIChangeLogServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIChangeLogService) EXPECT() *MockIChangeLogServiceMockRecorder {
	return m.recorder
}

// FindByID mocks base method.
func (m *MockIChangeLogService) FindByID(ctx context.Context, id string) (*changelog.Record, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindByID", ctx, id)
	ret0, _ := ret[0].(*changelog.Record)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindByID indicates an expected call of FindByID.
func (mr *MockIChangeLogServiceMockRecorder) FindByID(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindByID", reflect.TypeOf((*MockIChangeLogService)(nil).FindByID), ctx, id)
}

// FindByIDs mocks base method.
func (m *MockIChangeLogService) FindByIDs(ctx context.Context, ids []string, preserveOrder bool) ([]*changelog.Record, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindByIDs", ctx, ids, preserveOrder)
	ret0, _ := ret[0].([]*changelog.Record)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindByIDs indicates an expected call of FindByIDs.
func (mr *MockIChangeLogServiceMockRecorder) FindByIDs(ctx, ids, preserveOrder interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindByIDs", reflect.TypeOf((*MockIChangeLogService)(nil).FindByIDs), ctx, ids, preserveOrder)
}

// GetRecords mocks base method.
func (m *MockIChangeLogService) GetRecords(ctx context.Context, params *changelog.RecordsParams) (*changelog.Records, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRecords", ctx, params)
	ret0, _ := ret[0].(*changelog.Records)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRecords indicates an expected call of GetRecords.
func (mr *MockIChangeLogServiceMockRecorder) GetRecords(ctx, params interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRecords", reflect.TypeOf((*MockIChangeLogService)(nil).GetRecords), ctx, params)
}

// MockIFilter is a mock of IFilter interface.
type MockIFilter struct {
	ctrl     *gomock.Controller
	recorder *MockIFilterMockRecorder
}

// MockIFilterMockRecorder is the mock recorder for MockIFilter.
type MockIFilterMockRecorder struct {
	mock *MockIFilter
}

// NewMockIFilter creates a new mock instance.
func NewMockIFilter(ctrl *gomock.Controller) *MockIFilter {
	mock := &MockIFilter{ctrl: ctrl}
	mock.recorder = &MockIFilterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIFilter) EXPECT() *MockIFilterMockRecorder {
	return m.recorder
}

// Filter mocks base method.
func (m *MockIFilter) Filter(f interface{}) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Filter", f)
}

// Filter indicates an expected call of Filter.
func (mr *MockIFilterMockRecorder) Filter(f interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Filter", reflect.TypeOf((*MockIFilter)(nil).Filter), f)
}
