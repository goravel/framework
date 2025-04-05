// Code generated by mockery. DO NOT EDIT.

package packages

import (
	dst "github.com/dave/dst"
	dstutil "github.com/dave/dst/dstutil"

	mock "github.com/stretchr/testify/mock"
)

// GoNodeMatcher is an autogenerated mock type for the GoNodeMatcher type
type GoNodeMatcher struct {
	mock.Mock
}

type GoNodeMatcher_Expecter struct {
	mock *mock.Mock
}

func (_m *GoNodeMatcher) EXPECT() *GoNodeMatcher_Expecter {
	return &GoNodeMatcher_Expecter{mock: &_m.Mock}
}

// MatchCursor provides a mock function with given fields: cursor
func (_m *GoNodeMatcher) MatchCursor(cursor *dstutil.Cursor) bool {
	ret := _m.Called(cursor)

	if len(ret) == 0 {
		panic("no return value specified for MatchCursor")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func(*dstutil.Cursor) bool); ok {
		r0 = rf(cursor)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// GoNodeMatcher_MatchCursor_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'MatchCursor'
type GoNodeMatcher_MatchCursor_Call struct {
	*mock.Call
}

// MatchCursor is a helper method to define mock.On call
//   - cursor *dstutil.Cursor
func (_e *GoNodeMatcher_Expecter) MatchCursor(cursor interface{}) *GoNodeMatcher_MatchCursor_Call {
	return &GoNodeMatcher_MatchCursor_Call{Call: _e.mock.On("MatchCursor", cursor)}
}

func (_c *GoNodeMatcher_MatchCursor_Call) Run(run func(cursor *dstutil.Cursor)) *GoNodeMatcher_MatchCursor_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*dstutil.Cursor))
	})
	return _c
}

func (_c *GoNodeMatcher_MatchCursor_Call) Return(_a0 bool) *GoNodeMatcher_MatchCursor_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *GoNodeMatcher_MatchCursor_Call) RunAndReturn(run func(*dstutil.Cursor) bool) *GoNodeMatcher_MatchCursor_Call {
	_c.Call.Return(run)
	return _c
}

// MatchNode provides a mock function with given fields: node
func (_m *GoNodeMatcher) MatchNode(node dst.Node) bool {
	ret := _m.Called(node)

	if len(ret) == 0 {
		panic("no return value specified for MatchNode")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func(dst.Node) bool); ok {
		r0 = rf(node)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// GoNodeMatcher_MatchNode_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'MatchNode'
type GoNodeMatcher_MatchNode_Call struct {
	*mock.Call
}

// MatchNode is a helper method to define mock.On call
//   - node dst.Node
func (_e *GoNodeMatcher_Expecter) MatchNode(node interface{}) *GoNodeMatcher_MatchNode_Call {
	return &GoNodeMatcher_MatchNode_Call{Call: _e.mock.On("MatchNode", node)}
}

func (_c *GoNodeMatcher_MatchNode_Call) Run(run func(node dst.Node)) *GoNodeMatcher_MatchNode_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(dst.Node))
	})
	return _c
}

func (_c *GoNodeMatcher_MatchNode_Call) Return(_a0 bool) *GoNodeMatcher_MatchNode_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *GoNodeMatcher_MatchNode_Call) RunAndReturn(run func(dst.Node) bool) *GoNodeMatcher_MatchNode_Call {
	_c.Call.Return(run)
	return _c
}

// NewGoNodeMatcher creates a new instance of GoNodeMatcher. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewGoNodeMatcher(t interface {
	mock.TestingT
	Cleanup(func())
}) *GoNodeMatcher {
	mock := &GoNodeMatcher{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
