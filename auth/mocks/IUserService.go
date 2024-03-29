// Code generated by mockery v2.33.1. DO NOT EDIT.

package mocks

import (
	context "context"

	auth "github.com/bldsoft/gost/auth"

	mock "github.com/stretchr/testify/mock"
)

// IUserService is an autogenerated mock type for the IUserService type
type IUserService[PT auth.AuthenticablePtr[T], T interface{}] struct {
	mock.Mock
}

type IUserService_Expecter[PT auth.AuthenticablePtr[T], T interface{}] struct {
	mock *mock.Mock
}

func (_m *IUserService[PT, T]) EXPECT() *IUserService_Expecter[PT, T] {
	return &IUserService_Expecter[PT, T]{mock: &_m.Mock}
}

// Create provides a mock function with given fields: ctx, user, recoverDeleted
func (_m *IUserService[PT, T]) Create(ctx context.Context, user PT, recoverDeleted bool) error {
	ret := _m.Called(ctx, user, recoverDeleted)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, PT, bool) error); ok {
		r0 = rf(ctx, user, recoverDeleted)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// IUserService_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type IUserService_Create_Call[PT auth.AuthenticablePtr[T], T interface{}] struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - ctx context.Context
//   - user PT
//   - recoverDeleted bool
func (_e *IUserService_Expecter[PT, T]) Create(ctx interface{}, user interface{}, recoverDeleted interface{}) *IUserService_Create_Call[PT, T] {
	return &IUserService_Create_Call[PT, T]{Call: _e.mock.On("Create", ctx, user, recoverDeleted)}
}

func (_c *IUserService_Create_Call[PT, T]) Run(run func(ctx context.Context, user PT, recoverDeleted bool)) *IUserService_Create_Call[PT, T] {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(PT), args[2].(bool))
	})
	return _c
}

func (_c *IUserService_Create_Call[PT, T]) Return(_a0 error) *IUserService_Create_Call[PT, T] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *IUserService_Create_Call[PT, T]) RunAndReturn(run func(context.Context, PT, bool) error) *IUserService_Create_Call[PT, T] {
	_c.Call.Return(run)
	return _c
}

// Delete provides a mock function with given fields: ctx, id, archived
func (_m *IUserService[PT, T]) Delete(ctx context.Context, id string, archived bool) error {
	ret := _m.Called(ctx, id, archived)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, bool) error); ok {
		r0 = rf(ctx, id, archived)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// IUserService_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type IUserService_Delete_Call[PT auth.AuthenticablePtr[T], T interface{}] struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
//   - archived bool
func (_e *IUserService_Expecter[PT, T]) Delete(ctx interface{}, id interface{}, archived interface{}) *IUserService_Delete_Call[PT, T] {
	return &IUserService_Delete_Call[PT, T]{Call: _e.mock.On("Delete", ctx, id, archived)}
}

func (_c *IUserService_Delete_Call[PT, T]) Run(run func(ctx context.Context, id string, archived bool)) *IUserService_Delete_Call[PT, T] {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(bool))
	})
	return _c
}

func (_c *IUserService_Delete_Call[PT, T]) Return(_a0 error) *IUserService_Delete_Call[PT, T] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *IUserService_Delete_Call[PT, T]) RunAndReturn(run func(context.Context, string, bool) error) *IUserService_Delete_Call[PT, T] {
	_c.Call.Return(run)
	return _c
}

// GetAll provides a mock function with given fields: ctx, archived
func (_m *IUserService[PT, T]) GetAll(ctx context.Context, archived bool) ([]PT, error) {
	ret := _m.Called(ctx, archived)

	var r0 []PT
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, bool) ([]PT, error)); ok {
		return rf(ctx, archived)
	}
	if rf, ok := ret.Get(0).(func(context.Context, bool) []PT); ok {
		r0 = rf(ctx, archived)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]PT)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, bool) error); ok {
		r1 = rf(ctx, archived)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IUserService_GetAll_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAll'
type IUserService_GetAll_Call[PT auth.AuthenticablePtr[T], T interface{}] struct {
	*mock.Call
}

// GetAll is a helper method to define mock.On call
//   - ctx context.Context
//   - archived bool
func (_e *IUserService_Expecter[PT, T]) GetAll(ctx interface{}, archived interface{}) *IUserService_GetAll_Call[PT, T] {
	return &IUserService_GetAll_Call[PT, T]{Call: _e.mock.On("GetAll", ctx, archived)}
}

func (_c *IUserService_GetAll_Call[PT, T]) Run(run func(ctx context.Context, archived bool)) *IUserService_GetAll_Call[PT, T] {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(bool))
	})
	return _c
}

func (_c *IUserService_GetAll_Call[PT, T]) Return(_a0 []PT, _a1 error) *IUserService_GetAll_Call[PT, T] {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *IUserService_GetAll_Call[PT, T]) RunAndReturn(run func(context.Context, bool) ([]PT, error)) *IUserService_GetAll_Call[PT, T] {
	_c.Call.Return(run)
	return _c
}

// GetByID provides a mock function with given fields: ctx, id
func (_m *IUserService[PT, T]) GetByID(ctx context.Context, id string) (PT, error) {
	ret := _m.Called(ctx, id)

	var r0 PT
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (PT, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) PT); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Get(0).(PT)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IUserService_GetByID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetByID'
type IUserService_GetByID_Call[PT auth.AuthenticablePtr[T], T interface{}] struct {
	*mock.Call
}

// GetByID is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
func (_e *IUserService_Expecter[PT, T]) GetByID(ctx interface{}, id interface{}) *IUserService_GetByID_Call[PT, T] {
	return &IUserService_GetByID_Call[PT, T]{Call: _e.mock.On("GetByID", ctx, id)}
}

func (_c *IUserService_GetByID_Call[PT, T]) Run(run func(ctx context.Context, id string)) *IUserService_GetByID_Call[PT, T] {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *IUserService_GetByID_Call[PT, T]) Return(_a0 PT, _a1 error) *IUserService_GetByID_Call[PT, T] {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *IUserService_GetByID_Call[PT, T]) RunAndReturn(run func(context.Context, string) (PT, error)) *IUserService_GetByID_Call[PT, T] {
	_c.Call.Return(run)
	return _c
}

// Update provides a mock function with given fields: ctx, user
func (_m *IUserService[PT, T]) Update(ctx context.Context, user PT) error {
	ret := _m.Called(ctx, user)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, PT) error); ok {
		r0 = rf(ctx, user)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// IUserService_Update_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Update'
type IUserService_Update_Call[PT auth.AuthenticablePtr[T], T interface{}] struct {
	*mock.Call
}

// Update is a helper method to define mock.On call
//   - ctx context.Context
//   - user PT
func (_e *IUserService_Expecter[PT, T]) Update(ctx interface{}, user interface{}) *IUserService_Update_Call[PT, T] {
	return &IUserService_Update_Call[PT, T]{Call: _e.mock.On("Update", ctx, user)}
}

func (_c *IUserService_Update_Call[PT, T]) Run(run func(ctx context.Context, user PT)) *IUserService_Update_Call[PT, T] {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(PT))
	})
	return _c
}

func (_c *IUserService_Update_Call[PT, T]) Return(_a0 error) *IUserService_Update_Call[PT, T] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *IUserService_Update_Call[PT, T]) RunAndReturn(run func(context.Context, PT) error) *IUserService_Update_Call[PT, T] {
	_c.Call.Return(run)
	return _c
}

// UpdatePassword provides a mock function with given fields: ctx, id, password
func (_m *IUserService[PT, T]) UpdatePassword(ctx context.Context, id string, password string) error {
	ret := _m.Called(ctx, id, password)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, id, password)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// IUserService_UpdatePassword_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdatePassword'
type IUserService_UpdatePassword_Call[PT auth.AuthenticablePtr[T], T interface{}] struct {
	*mock.Call
}

// UpdatePassword is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
//   - password string
func (_e *IUserService_Expecter[PT, T]) UpdatePassword(ctx interface{}, id interface{}, password interface{}) *IUserService_UpdatePassword_Call[PT, T] {
	return &IUserService_UpdatePassword_Call[PT, T]{Call: _e.mock.On("UpdatePassword", ctx, id, password)}
}

func (_c *IUserService_UpdatePassword_Call[PT, T]) Run(run func(ctx context.Context, id string, password string)) *IUserService_UpdatePassword_Call[PT, T] {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *IUserService_UpdatePassword_Call[PT, T]) Return(_a0 error) *IUserService_UpdatePassword_Call[PT, T] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *IUserService_UpdatePassword_Call[PT, T]) RunAndReturn(run func(context.Context, string, string) error) *IUserService_UpdatePassword_Call[PT, T] {
	_c.Call.Return(run)
	return _c
}

// NewIUserService creates a new instance of IUserService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewIUserService[PT auth.AuthenticablePtr[T], T interface{}](t interface {
	mock.TestingT
	Cleanup(func())
}) *IUserService[PT, T] {
	mock := &IUserService[PT, T]{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
