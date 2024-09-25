// Code generated by mockery v2.36.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// IEntityArchived is an autogenerated mock type for the IEntityArchived type
type IEntityArchived struct {
	mock.Mock
}

type IEntityArchived_Expecter struct {
	mock *mock.Mock
}

func (_m *IEntityArchived) EXPECT() *IEntityArchived_Expecter {
	return &IEntityArchived_Expecter{mock: &_m.Mock}
}

// IsArchived provides a mock function with given fields:
func (_m *IEntityArchived) IsArchived() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// IEntityArchived_IsArchived_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsArchived'
type IEntityArchived_IsArchived_Call struct {
	*mock.Call
}

// IsArchived is a helper method to define mock.On call
func (_e *IEntityArchived_Expecter) IsArchived() *IEntityArchived_IsArchived_Call {
	return &IEntityArchived_IsArchived_Call{Call: _e.mock.On("IsArchived")}
}

func (_c *IEntityArchived_IsArchived_Call) Run(run func()) *IEntityArchived_IsArchived_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *IEntityArchived_IsArchived_Call) Return(_a0 bool) *IEntityArchived_IsArchived_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *IEntityArchived_IsArchived_Call) RunAndReturn(run func() bool) *IEntityArchived_IsArchived_Call {
	_c.Call.Return(run)
	return _c
}

// NewIEntityArchived creates a new instance of IEntityArchived. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewIEntityArchived(t interface {
	mock.TestingT
	Cleanup(func())
}) *IEntityArchived {
	mock := &IEntityArchived{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}