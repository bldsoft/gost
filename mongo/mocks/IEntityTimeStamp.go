// Code generated by mockery v2.36.0. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"

	time "time"
)

// IEntityTimeStamp is an autogenerated mock type for the IEntityTimeStamp type
type IEntityTimeStamp struct {
	mock.Mock
}

type IEntityTimeStamp_Expecter struct {
	mock *mock.Mock
}

func (_m *IEntityTimeStamp) EXPECT() *IEntityTimeStamp_Expecter {
	return &IEntityTimeStamp_Expecter{mock: &_m.Mock}
}

// SetCreateFields provides a mock function with given fields: createTime, createUserID
func (_m *IEntityTimeStamp) SetCreateFields(createTime time.Time, createUserID interface{}) {
	_m.Called(createTime, createUserID)
}

// IEntityTimeStamp_SetCreateFields_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetCreateFields'
type IEntityTimeStamp_SetCreateFields_Call struct {
	*mock.Call
}

// SetCreateFields is a helper method to define mock.On call
//   - createTime time.Time
//   - createUserID interface{}
func (_e *IEntityTimeStamp_Expecter) SetCreateFields(createTime interface{}, createUserID interface{}) *IEntityTimeStamp_SetCreateFields_Call {
	return &IEntityTimeStamp_SetCreateFields_Call{Call: _e.mock.On("SetCreateFields", createTime, createUserID)}
}

func (_c *IEntityTimeStamp_SetCreateFields_Call) Run(run func(createTime time.Time, createUserID interface{})) *IEntityTimeStamp_SetCreateFields_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(time.Time), args[1].(interface{}))
	})
	return _c
}

func (_c *IEntityTimeStamp_SetCreateFields_Call) Return() *IEntityTimeStamp_SetCreateFields_Call {
	_c.Call.Return()
	return _c
}

func (_c *IEntityTimeStamp_SetCreateFields_Call) RunAndReturn(run func(time.Time, interface{})) *IEntityTimeStamp_SetCreateFields_Call {
	_c.Call.Return(run)
	return _c
}

// SetUpdateFields provides a mock function with given fields: cupdateTime, updateUserID
func (_m *IEntityTimeStamp) SetUpdateFields(cupdateTime time.Time, updateUserID interface{}) {
	_m.Called(cupdateTime, updateUserID)
}

// IEntityTimeStamp_SetUpdateFields_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetUpdateFields'
type IEntityTimeStamp_SetUpdateFields_Call struct {
	*mock.Call
}

// SetUpdateFields is a helper method to define mock.On call
//   - cupdateTime time.Time
//   - updateUserID interface{}
func (_e *IEntityTimeStamp_Expecter) SetUpdateFields(cupdateTime interface{}, updateUserID interface{}) *IEntityTimeStamp_SetUpdateFields_Call {
	return &IEntityTimeStamp_SetUpdateFields_Call{Call: _e.mock.On("SetUpdateFields", cupdateTime, updateUserID)}
}

func (_c *IEntityTimeStamp_SetUpdateFields_Call) Run(run func(cupdateTime time.Time, updateUserID interface{})) *IEntityTimeStamp_SetUpdateFields_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(time.Time), args[1].(interface{}))
	})
	return _c
}

func (_c *IEntityTimeStamp_SetUpdateFields_Call) Return() *IEntityTimeStamp_SetUpdateFields_Call {
	_c.Call.Return()
	return _c
}

func (_c *IEntityTimeStamp_SetUpdateFields_Call) RunAndReturn(run func(time.Time, interface{})) *IEntityTimeStamp_SetUpdateFields_Call {
	_c.Call.Return(run)
	return _c
}

// NewIEntityTimeStamp creates a new instance of IEntityTimeStamp. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewIEntityTimeStamp(t interface {
	mock.TestingT
	Cleanup(func())
}) *IEntityTimeStamp {
	mock := &IEntityTimeStamp{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
