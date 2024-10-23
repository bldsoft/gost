// Code generated by mockery v2.36.0. DO NOT EDIT.

package mocks

import (
	time "time"

	mock "github.com/stretchr/testify/mock"
)

// Repository is an autogenerated mock type for the Repository type
type Repository[T interface{}] struct {
	mock.Mock
}

type Repository_Expecter[T interface{}] struct {
	mock *mock.Mock
}

func (_m *Repository[T]) EXPECT() *Repository_Expecter[T] {
	return &Repository_Expecter[T]{mock: &_m.Mock}
}

// Delete provides a mock function with given fields: key
func (_m *Repository[T]) Delete(key string) error {
	ret := _m.Called(key)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(key)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Repository_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type Repository_Delete_Call[T interface{}] struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - key string
func (_e *Repository_Expecter[T]) Delete(key interface{}) *Repository_Delete_Call[T] {
	return &Repository_Delete_Call[T]{Call: _e.mock.On("Delete", key)}
}

func (_c *Repository_Delete_Call[T]) Run(run func(key string)) *Repository_Delete_Call[T] {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Repository_Delete_Call[T]) Return(_a0 error) *Repository_Delete_Call[T] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Repository_Delete_Call[T]) RunAndReturn(run func(string) error) *Repository_Delete_Call[T] {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: key
func (_m *Repository[T]) Get(key string) (T, error) {
	ret := _m.Called(key)

	var r0 T
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (T, error)); ok {
		return rf(key)
	}
	if rf, ok := ret.Get(0).(func(string) T); ok {
		r0 = rf(key)
	} else {
		r0 = ret.Get(0).(T)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Repository_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type Repository_Get_Call[T interface{}] struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - key string
func (_e *Repository_Expecter[T]) Get(key interface{}) *Repository_Get_Call[T] {
	return &Repository_Get_Call[T]{Call: _e.mock.On("Get", key)}
}

func (_c *Repository_Get_Call[T]) Run(run func(key string)) *Repository_Get_Call[T] {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Repository_Get_Call[T]) Return(_a0 T, _a1 error) *Repository_Get_Call[T] {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Repository_Get_Call[T]) RunAndReturn(run func(string) (T, error)) *Repository_Get_Call[T] {
	_c.Call.Return(run)
	return _c
}

// Reset provides a mock function with given fields:
func (_m *Repository[T]) Reset() {
	_m.Called()
}

// Repository_Reset_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Reset'
type Repository_Reset_Call[T interface{}] struct {
	*mock.Call
}

// Reset is a helper method to define mock.On call
func (_e *Repository_Expecter[T]) Reset() *Repository_Reset_Call[T] {
	return &Repository_Reset_Call[T]{Call: _e.mock.On("Reset")}
}

func (_c *Repository_Reset_Call[T]) Run(run func()) *Repository_Reset_Call[T] {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Repository_Reset_Call[T]) Return() *Repository_Reset_Call[T] {
	_c.Call.Return()
	return _c
}

func (_c *Repository_Reset_Call[T]) RunAndReturn(run func()) *Repository_Reset_Call[T] {
	_c.Call.Return(run)
	return _c
}

// Set provides a mock function with given fields: key, value
func (_m *Repository[T]) Set(key string, value T) error {
	ret := _m.Called(key, value)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, T) error); ok {
		r0 = rf(key, value)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Repository_Set_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Set'
type Repository_Set_Call[T interface{}] struct {
	*mock.Call
}

// Set is a helper method to define mock.On call
//   - key string
//   - value T
func (_e *Repository_Expecter[T]) Set(key interface{}, value interface{}) *Repository_Set_Call[T] {
	return &Repository_Set_Call[T]{Call: _e.mock.On("Set", key, value)}
}

func (_c *Repository_Set_Call[T]) Run(run func(key string, value T)) *Repository_Set_Call[T] {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(T))
	})
	return _c
}

func (_c *Repository_Set_Call[T]) Return(_a0 error) *Repository_Set_Call[T] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Repository_Set_Call[T]) RunAndReturn(run func(string, T) error) *Repository_Set_Call[T] {
	_c.Call.Return(run)
	return _c
}

// SetFor provides a mock function with given fields: key, value, ttl
func (_m *Repository[T]) SetFor(key string, value T, ttl time.Duration) error {
	ret := _m.Called(key, value, ttl)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, T, time.Duration) error); ok {
		r0 = rf(key, value, ttl)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Repository_SetFor_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetFor'
type Repository_SetFor_Call[T interface{}] struct {
	*mock.Call
}

// SetFor is a helper method to define mock.On call
//   - key string
//   - value T
//   - ttl time.Duration
func (_e *Repository_Expecter[T]) SetFor(key interface{}, value interface{}, ttl interface{}) *Repository_SetFor_Call[T] {
	return &Repository_SetFor_Call[T]{Call: _e.mock.On("SetFor", key, value, ttl)}
}

func (_c *Repository_SetFor_Call[T]) Run(run func(key string, value T, ttl time.Duration)) *Repository_SetFor_Call[T] {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(T), args[2].(time.Duration))
	})
	return _c
}

func (_c *Repository_SetFor_Call[T]) Return(_a0 error) *Repository_SetFor_Call[T] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Repository_SetFor_Call[T]) RunAndReturn(run func(string, T, time.Duration) error) *Repository_SetFor_Call[T] {
	_c.Call.Return(run)
	return _c
}

// NewRepository creates a new instance of Repository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewRepository[T interface{}](t interface {
	mock.TestingT
	Cleanup(func())
}) *Repository[T] {
	mock := &Repository[T]{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}