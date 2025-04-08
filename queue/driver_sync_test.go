package queue

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/queue"
	mocksqueue "github.com/goravel/framework/mocks/queue"
)

var (
	testJobOne []any
	testJobTwo []any
)

type SyncTestSuite struct {
	suite.Suite
	app        *Application
	mockConfig *mocksqueue.Config
}

func TestSyncTestSuite(t *testing.T) {
	suite.Run(t, new(SyncTestSuite))
}

func (s *SyncTestSuite) SetupSuite() {
	s.mockConfig = mocksqueue.NewConfig(s.T())

	s.app = &Application{
		config: s.mockConfig,
		job:    NewJobRepository(),
	}

	s.app.Register([]queue.Job{&TestJobOne{}, &TestJobTwo{}, &TestJobErr{}})
}

func (s *SyncTestSuite) SetupTest() {
	testJobOne = nil
	testJobTwo = nil

	s.mockConfig.EXPECT().Default().Return("sync", "default", 1).Once()
	s.mockConfig.EXPECT().QueueKey("sync", "default").Return("sync_queue").Once()
	s.mockConfig.EXPECT().Driver("sync").Return(queue.DriverSync).Once()
}

func (s *SyncTestSuite) TestDelay() {
	args := []queue.Arg{
		{
			Type:  "string",
			Value: "a",
		},
		{
			Type:  "int",
			Value: 1,
		},
		{
			Type:  "[]string",
			Value: []string{"b", "c"},
		},
		{
			Type:  "[]int",
			Value: []int{1, 2, 3},
		},
	}
	s.Nil(s.app.Job(&TestJobOne{}, args).Delay(time.Now().Add(time.Second)).Dispatch())
	s.Equal([]any{"a", 1, []string{"b", "c"}, []int{1, 2, 3}}, testJobOne)
}

func (s *SyncTestSuite) TestDispatch() {
	args := []queue.Arg{
		{
			Type:  "string",
			Value: "a",
		},
		{
			Type:  "int",
			Value: 1,
		},
		{
			Type:  "[]string",
			Value: []string{"b", "c"},
		},
		{
			Type:  "[]int",
			Value: []int{1, 2, 3},
		},
	}

	s.Nil(s.app.Job(&TestJobOne{}, args).Dispatch())
	s.Equal([]any{"a", 1, []string{"b", "c"}, []int{1, 2, 3}}, testJobOne)
}

func (s *SyncTestSuite) TestChainDispatch() {
	argsOne := []queue.Arg{
		{
			Type:  "string",
			Value: "a",
		},
		{
			Type:  "int",
			Value: 1,
		},
		{
			Type:  "[]string",
			Value: []string{"b", "c"},
		},
		{
			Type:  "[]int",
			Value: []int{1, 2, 3},
		},
	}
	argsTwo := []queue.Arg{
		{
			Type:  "string",
			Value: "a",
		},
		{
			Type:  "int",
			Value: 2,
		},
		{
			Type:  "[]string",
			Value: []string{"d", "f"},
		},
		{
			Type:  "[]int",
			Value: []int{4, 5, 6},
		},
	}
	s.Nil(s.app.Chain([]queue.Jobs{
		{
			Job:  &TestJobOne{},
			Args: argsOne,
		},
		{
			Job:  &TestJobTwo{},
			Args: argsTwo,
		},
	}).Dispatch())

	s.Equal([]any{"a", 1, []string{"b", "c"}, []int{1, 2, 3}}, testJobOne)
	s.Equal([]any{"a", 2, []string{"d", "f"}, []int{4, 5, 6}}, testJobTwo)
}

func (s *SyncTestSuite) TestChainDispatchWithError() {
	argsOne := []queue.Arg{
		{
			Type:  "string",
			Value: "a",
		},
		{
			Type:  "int",
			Value: 1,
		},
		{
			Type:  "[]string",
			Value: []string{"b", "c"},
		},
		{
			Type:  "[]int",
			Value: []int{1, 2, 3},
		},
	}
	argsTwo := []queue.Arg{
		{
			Type:  "string",
			Value: "a",
		},
		{
			Type:  "int",
			Value: 2,
		},
		{
			Type:  "[]string",
			Value: []string{"d", "f"},
		},
		{
			Type:  "[]int",
			Value: []int{4, 5, 6},
		},
	}

	s.Equal(assert.AnError, s.app.Chain([]queue.Jobs{
		{
			Job:  &TestJobOne{},
			Args: argsOne,
		},
		{
			Job: &TestJobErr{},
		},
		{
			Job:  &TestJobTwo{},
			Args: argsTwo,
		},
	}).Dispatch())

	s.Equal([]any{"a", 1, []string{"b", "c"}, []int{1, 2, 3}}, testJobOne)
	s.Nil(testJobTwo)
}

type TestJobOne struct {
}

// Signature The name and signature of the job.
func (r *TestJobOne) Signature() string {
	return "test_job_one"
}

// Handle Execute the job.
func (r *TestJobOne) Handle(args ...any) error {
	testJobOne = args

	return nil
}

type TestJobTwo struct {
}

// Signature The name and signature of the job.
func (r *TestJobTwo) Signature() string {
	return "test_job_two"
}

// Handle Execute the job.
func (r *TestJobTwo) Handle(args ...any) error {
	testJobTwo = args

	return nil
}

type TestJobErr struct {
}

// Signature The name and signature of the job.
func (r *TestJobErr) Signature() string {
	return "test_job_err"
}

// Handle Execute the job.
func (r *TestJobErr) Handle(args ...any) error {
	return assert.AnError
}
