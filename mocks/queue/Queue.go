// Code generated by mockery. DO NOT EDIT.

package queue

import (
	queue "github.com/goravel/framework/contracts/queue"
	mock "github.com/stretchr/testify/mock"
)

// Queue is an autogenerated mock type for the Queue type
type Queue struct {
	mock.Mock
}

type Queue_Expecter struct {
	mock *mock.Mock
}

func (_m *Queue) EXPECT() *Queue_Expecter {
	return &Queue_Expecter{mock: &_m.Mock}
}

// Chain provides a mock function with given fields: jobs
func (_m *Queue) Chain(jobs []queue.Jobs) queue.Task {
	ret := _m.Called(jobs)

	if len(ret) == 0 {
		panic("no return value specified for Chain")
	}

	var r0 queue.Task
	if rf, ok := ret.Get(0).(func([]queue.Jobs) queue.Task); ok {
		r0 = rf(jobs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(queue.Task)
		}
	}

	return r0
}

// Queue_Chain_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Chain'
type Queue_Chain_Call struct {
	*mock.Call
}

// Chain is a helper method to define mock.On call
//   - jobs []queue.Jobs
func (_e *Queue_Expecter) Chain(jobs interface{}) *Queue_Chain_Call {
	return &Queue_Chain_Call{Call: _e.mock.On("Chain", jobs)}
}

func (_c *Queue_Chain_Call) Run(run func(jobs []queue.Jobs)) *Queue_Chain_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]queue.Jobs))
	})
	return _c
}

func (_c *Queue_Chain_Call) Return(_a0 queue.Task) *Queue_Chain_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Queue_Chain_Call) RunAndReturn(run func([]queue.Jobs) queue.Task) *Queue_Chain_Call {
	_c.Call.Return(run)
	return _c
}

// GetJobs provides a mock function with given fields:
func (_m *Queue) GetJobs() []queue.Job {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetJobs")
	}

	var r0 []queue.Job
	if rf, ok := ret.Get(0).(func() []queue.Job); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]queue.Job)
		}
	}

	return r0
}

// Queue_GetJobs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetJobs'
type Queue_GetJobs_Call struct {
	*mock.Call
}

// GetJobs is a helper method to define mock.On call
func (_e *Queue_Expecter) GetJobs() *Queue_GetJobs_Call {
	return &Queue_GetJobs_Call{Call: _e.mock.On("GetJobs")}
}

func (_c *Queue_GetJobs_Call) Run(run func()) *Queue_GetJobs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Queue_GetJobs_Call) Return(_a0 []queue.Job) *Queue_GetJobs_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Queue_GetJobs_Call) RunAndReturn(run func() []queue.Job) *Queue_GetJobs_Call {
	_c.Call.Return(run)
	return _c
}

// Job provides a mock function with given fields: job, args
func (_m *Queue) Job(job queue.Job, args []queue.Arg) queue.Task {
	ret := _m.Called(job, args)

	if len(ret) == 0 {
		panic("no return value specified for Job")
	}

	var r0 queue.Task
	if rf, ok := ret.Get(0).(func(queue.Job, []queue.Arg) queue.Task); ok {
		r0 = rf(job, args)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(queue.Task)
		}
	}

	return r0
}

// Queue_Job_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Job'
type Queue_Job_Call struct {
	*mock.Call
}

// Job is a helper method to define mock.On call
//   - job queue.Job
//   - args []queue.Arg
func (_e *Queue_Expecter) Job(job interface{}, args interface{}) *Queue_Job_Call {
	return &Queue_Job_Call{Call: _e.mock.On("Job", job, args)}
}

func (_c *Queue_Job_Call) Run(run func(job queue.Job, args []queue.Arg)) *Queue_Job_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(queue.Job), args[1].([]queue.Arg))
	})
	return _c
}

func (_c *Queue_Job_Call) Return(_a0 queue.Task) *Queue_Job_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Queue_Job_Call) RunAndReturn(run func(queue.Job, []queue.Arg) queue.Task) *Queue_Job_Call {
	_c.Call.Return(run)
	return _c
}

// Register provides a mock function with given fields: jobs
func (_m *Queue) Register(jobs []queue.Job) {
	_m.Called(jobs)
}

// Queue_Register_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Register'
type Queue_Register_Call struct {
	*mock.Call
}

// Register is a helper method to define mock.On call
//   - jobs []queue.Job
func (_e *Queue_Expecter) Register(jobs interface{}) *Queue_Register_Call {
	return &Queue_Register_Call{Call: _e.mock.On("Register", jobs)}
}

func (_c *Queue_Register_Call) Run(run func(jobs []queue.Job)) *Queue_Register_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]queue.Job))
	})
	return _c
}

func (_c *Queue_Register_Call) Return() *Queue_Register_Call {
	_c.Call.Return()
	return _c
}

func (_c *Queue_Register_Call) RunAndReturn(run func([]queue.Job)) *Queue_Register_Call {
	_c.Call.Return(run)
	return _c
}

// Worker provides a mock function with given fields: args
func (_m *Queue) Worker(args *queue.Args) queue.Worker {
	ret := _m.Called(args)

	if len(ret) == 0 {
		panic("no return value specified for Worker")
	}

	var r0 queue.Worker
	if rf, ok := ret.Get(0).(func(*queue.Args) queue.Worker); ok {
		r0 = rf(args)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(queue.Worker)
		}
	}

	return r0
}

// Queue_Worker_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Worker'
type Queue_Worker_Call struct {
	*mock.Call
}

// Worker is a helper method to define mock.On call
//   - args *queue.Args
func (_e *Queue_Expecter) Worker(args interface{}) *Queue_Worker_Call {
	return &Queue_Worker_Call{Call: _e.mock.On("Worker", args)}
}

func (_c *Queue_Worker_Call) Run(run func(args *queue.Args)) *Queue_Worker_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*queue.Args))
	})
	return _c
}

func (_c *Queue_Worker_Call) Return(_a0 queue.Worker) *Queue_Worker_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Queue_Worker_Call) RunAndReturn(run func(*queue.Args) queue.Worker) *Queue_Worker_Call {
	_c.Call.Return(run)
	return _c
}

// NewQueue creates a new instance of Queue. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewQueue(t interface {
	mock.TestingT
	Cleanup(func())
}) *Queue {
	mock := &Queue{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
