// Code generated by MockGen. DO NOT EDIT.
// Source: mongo/interface.go

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	repository "github.com/bldsoft/gost/repository"
	gomock "github.com/golang/mock/gomock"
	mongo0 "go.mongodb.org/mongo-driver/mongo"
)

// MockRepository is a mock of Repository interface.
type MockRepository[T any, U repository.IEntityIDPtr[T]] struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryMockRecorder[T, U]
}

// MockRepositoryMockRecorder is the mock recorder for MockRepository.
type MockRepositoryMockRecorder[T any, U repository.IEntityIDPtr[T]] struct {
	mock *MockRepository[T, U]
}

// NewMockRepository creates a new mock instance.
func NewMockRepository[T any, U repository.IEntityIDPtr[T]](ctrl *gomock.Controller) *MockRepository[T, U] {
	mock := &MockRepository[T, U]{ctrl: ctrl}
	mock.recorder = &MockRepositoryMockRecorder[T, U]{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRepository[T, U]) EXPECT() *MockRepositoryMockRecorder[T, U] {
	return m.recorder
}

// AggregateOne mocks base method.
func (m *MockRepository[T, U]) AggregateOne(ctx context.Context, pipeline mongo0.Pipeline, entity interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AggregateOne", ctx, pipeline, entity)
	ret0, _ := ret[0].(error)
	return ret0
}

// AggregateOne indicates an expected call of AggregateOne.
func (mr *MockRepositoryMockRecorder[T, U]) AggregateOne(ctx, pipeline, entity interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AggregateOne", reflect.TypeOf((*MockRepository[T, U])(nil).AggregateOne), ctx, pipeline, entity)
}

// Collection mocks base method.
func (m *MockRepository[T, U]) Collection() *mongo0.Collection {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Collection")
	ret0, _ := ret[0].(*mongo0.Collection)
	return ret0
}

// Collection indicates an expected call of Collection.
func (mr *MockRepositoryMockRecorder[T, U]) Collection() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Collection", reflect.TypeOf((*MockRepository[T, U])(nil).Collection))
}

// Delete mocks base method.
func (m *MockRepository[T, U]) Delete(ctx context.Context, id interface{}, options ...*repository.QueryOptions) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, id}
	for _, a := range options {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Delete", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockRepositoryMockRecorder[T, U]) Delete(ctx, id interface{}, options ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, id}, options...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockRepository[T, U])(nil).Delete), varargs...)
}

// DeleteMany mocks base method.
func (m *MockRepository[T, U]) DeleteMany(ctx context.Context, filter interface{}, options ...*repository.QueryOptions) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, filter}
	for _, a := range options {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DeleteMany", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteMany indicates an expected call of DeleteMany.
func (mr *MockRepositoryMockRecorder[T, U]) DeleteMany(ctx, filter interface{}, options ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, filter}, options...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteMany", reflect.TypeOf((*MockRepository[T, U])(nil).DeleteMany), varargs...)
}

// Find mocks base method.
func (m *MockRepository[T, U]) Find(ctx context.Context, filter interface{}, opt ...*repository.QueryOptions) ([]U, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, filter}
	for _, a := range opt {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Find", varargs...)
	ret0, _ := ret[0].([]U)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Find indicates an expected call of Find.
func (mr *MockRepositoryMockRecorder[T, U]) Find(ctx, filter interface{}, opt ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, filter}, opt...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Find", reflect.TypeOf((*MockRepository[T, U])(nil).Find), varargs...)
}

// FindByID mocks base method.
func (m *MockRepository[T, U]) FindByID(ctx context.Context, id interface{}, options ...*repository.QueryOptions) (U, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, id}
	for _, a := range options {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "FindByID", varargs...)
	ret0, _ := ret[0].(U)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindByID indicates an expected call of FindByID.
func (mr *MockRepositoryMockRecorder[T, U]) FindByID(ctx, id interface{}, options ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, id}, options...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindByID", reflect.TypeOf((*MockRepository[T, U])(nil).FindByID), varargs...)
}

// FindByIDs mocks base method.
func (m *MockRepository[T, U]) FindByIDs(ctx context.Context, ids []interface{}, preserveOrder bool, options ...*repository.QueryOptions) ([]U, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, ids, preserveOrder}
	for _, a := range options {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "FindByIDs", varargs...)
	ret0, _ := ret[0].([]U)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindByIDs indicates an expected call of FindByIDs.
func (mr *MockRepositoryMockRecorder[T, U]) FindByIDs(ctx, ids, preserveOrder interface{}, options ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, ids, preserveOrder}, options...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindByIDs", reflect.TypeOf((*MockRepository[T, U])(nil).FindByIDs), varargs...)
}

// FindByStringIDs mocks base method.
func (m *MockRepository[T, U]) FindByStringIDs(ctx context.Context, ids []string, preserveOrder bool, options ...*repository.QueryOptions) ([]U, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, ids, preserveOrder}
	for _, a := range options {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "FindByStringIDs", varargs...)
	ret0, _ := ret[0].([]U)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindByStringIDs indicates an expected call of FindByStringIDs.
func (mr *MockRepositoryMockRecorder[T, U]) FindByStringIDs(ctx, ids, preserveOrder interface{}, options ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, ids, preserveOrder}, options...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindByStringIDs", reflect.TypeOf((*MockRepository[T, U])(nil).FindByStringIDs), varargs...)
}

// FindOne mocks base method.
func (m *MockRepository[T, U]) FindOne(ctx context.Context, filter interface{}, opts ...*repository.QueryOptions) (U, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, filter}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "FindOne", varargs...)
	ret0, _ := ret[0].(U)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindOne indicates an expected call of FindOne.
func (mr *MockRepositoryMockRecorder[T, U]) FindOne(ctx, filter interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, filter}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindOne", reflect.TypeOf((*MockRepository[T, U])(nil).FindOne), varargs...)
}

// GetAll mocks base method.
func (m *MockRepository[T, U]) GetAll(ctx context.Context, options ...*repository.QueryOptions) ([]U, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx}
	for _, a := range options {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetAll", varargs...)
	ret0, _ := ret[0].([]U)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAll indicates an expected call of GetAll.
func (mr *MockRepositoryMockRecorder[T, U]) GetAll(ctx interface{}, options ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx}, options...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAll", reflect.TypeOf((*MockRepository[T, U])(nil).GetAll), varargs...)
}

// Insert mocks base method.
func (m *MockRepository[T, U]) Insert(ctx context.Context, entity U) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Insert", ctx, entity)
	ret0, _ := ret[0].(error)
	return ret0
}

// Insert indicates an expected call of Insert.
func (mr *MockRepositoryMockRecorder[T, U]) Insert(ctx, entity interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Insert", reflect.TypeOf((*MockRepository[T, U])(nil).Insert), ctx, entity)
}

// InsertMany mocks base method.
func (m *MockRepository[T, U]) InsertMany(ctx context.Context, entities []U) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertMany", ctx, entities)
	ret0, _ := ret[0].(error)
	return ret0
}

// InsertMany indicates an expected call of InsertMany.
func (mr *MockRepositoryMockRecorder[T, U]) InsertMany(ctx, entities interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertMany", reflect.TypeOf((*MockRepository[T, U])(nil).InsertMany), ctx, entities)
}

// Name mocks base method.
func (m *MockRepository[T, U]) Name() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Name")
	ret0, _ := ret[0].(string)
	return ret0
}

// Name indicates an expected call of Name.
func (mr *MockRepositoryMockRecorder[T, U]) Name() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Name", reflect.TypeOf((*MockRepository[T, U])(nil).Name))
}

// Update mocks base method.
func (m *MockRepository[T, U]) Update(ctx context.Context, entity U, options ...*repository.QueryOptions) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, entity}
	for _, a := range options {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Update", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *MockRepositoryMockRecorder[T, U]) Update(ctx, entity interface{}, options ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, entity}, options...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockRepository[T, U])(nil).Update), varargs...)
}

// UpdateAndGetByID mocks base method.
func (m *MockRepository[T, U]) UpdateAndGetByID(ctx context.Context, updateEntity U, returnNewDocument bool, queryOpt ...*repository.QueryOptions) (U, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, updateEntity, returnNewDocument}
	for _, a := range queryOpt {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpdateAndGetByID", varargs...)
	ret0, _ := ret[0].(U)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateAndGetByID indicates an expected call of UpdateAndGetByID.
func (mr *MockRepositoryMockRecorder[T, U]) UpdateAndGetByID(ctx, updateEntity, returnNewDocument interface{}, queryOpt ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, updateEntity, returnNewDocument}, queryOpt...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateAndGetByID", reflect.TypeOf((*MockRepository[T, U])(nil).UpdateAndGetByID), varargs...)
}

// UpdateMany mocks base method.
func (m *MockRepository[T, U]) UpdateMany(ctx context.Context, entities []U) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateMany", ctx, entities)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateMany indicates an expected call of UpdateMany.
func (mr *MockRepositoryMockRecorder[T, U]) UpdateMany(ctx, entities interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateMany", reflect.TypeOf((*MockRepository[T, U])(nil).UpdateMany), ctx, entities)
}

// UpdateOne mocks base method.
func (m *MockRepository[T, U]) UpdateOne(ctx context.Context, filter, update interface{}, options ...*repository.QueryOptions) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, filter, update}
	for _, a := range options {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpdateOne", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateOne indicates an expected call of UpdateOne.
func (mr *MockRepositoryMockRecorder[T, U]) UpdateOne(ctx, filter, update interface{}, options ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, filter, update}, options...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateOne", reflect.TypeOf((*MockRepository[T, U])(nil).UpdateOne), varargs...)
}

// Upsert mocks base method.
func (m *MockRepository[T, U]) Upsert(ctx context.Context, entity U, opt ...*repository.QueryOptions) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, entity}
	for _, a := range opt {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Upsert", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Upsert indicates an expected call of Upsert.
func (mr *MockRepositoryMockRecorder[T, U]) Upsert(ctx, entity interface{}, opt ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, entity}, opt...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Upsert", reflect.TypeOf((*MockRepository[T, U])(nil).Upsert), varargs...)
}

// UpsertOne mocks base method.
func (m *MockRepository[T, U]) UpsertOne(ctx context.Context, filter interface{}, update U) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpsertOne", ctx, filter, update)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpsertOne indicates an expected call of UpsertOne.
func (mr *MockRepositoryMockRecorder[T, U]) UpsertOne(ctx, filter, update interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpsertOne", reflect.TypeOf((*MockRepository[T, U])(nil).UpsertOne), ctx, filter, update)
}

// WithTransaction mocks base method.
func (m *MockRepository[T, U]) WithTransaction(ctx context.Context, f func(mongo0.SessionContext) (interface{}, error)) (interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithTransaction", ctx, f)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WithTransaction indicates an expected call of WithTransaction.
func (mr *MockRepositoryMockRecorder[T, U]) WithTransaction(ctx, f interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithTransaction", reflect.TypeOf((*MockRepository[T, U])(nil).WithTransaction), ctx, f)
}