// Code generated by mockery v2.36.0. DO NOT EDIT.

package mocks

import (
	cache "github.com/bldsoft/gost/cache/v2"
	mock "github.com/stretchr/testify/mock"
)

// ItemF is an autogenerated mock type for the ItemF type
type ItemF struct {
	mock.Mock
}

type ItemF_Expecter struct {
	mock *mock.Mock
}

func (_m *ItemF) EXPECT() *ItemF_Expecter {
	return &ItemF_Expecter{mock: &_m.Mock}
}

// Execute provides a mock function with given fields: _a0
func (_m *ItemF) Execute(_a0 *cache.Item) {
	_m.Called(_a0)
}

// ItemF_Execute_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Execute'
type ItemF_Execute_Call struct {
	*mock.Call
}

// Execute is a helper method to define mock.On call
//   - _a0 *cache.Item
func (_e *ItemF_Expecter) Execute(_a0 interface{}) *ItemF_Execute_Call {
	return &ItemF_Execute_Call{Call: _e.mock.On("Execute", _a0)}
}

func (_c *ItemF_Execute_Call) Run(run func(_a0 *cache.Item)) *ItemF_Execute_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*cache.Item))
	})
	return _c
}

func (_c *ItemF_Execute_Call) Return() *ItemF_Execute_Call {
	_c.Call.Return()
	return _c
}

func (_c *ItemF_Execute_Call) RunAndReturn(run func(*cache.Item)) *ItemF_Execute_Call {
	_c.Call.Return(run)
	return _c
}

// NewItemF creates a new instance of ItemF. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewItemF(t interface {
	mock.TestingT
	Cleanup(func())
}) *ItemF {
	mock := &ItemF{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
