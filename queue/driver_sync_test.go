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
		config:    s.mockConfig,
		jobStorer: NewJobStorer(),
	}

	s.app.Register([]queue.Job{&TestJobOne{}, &TestJobTwo{}, &TestJobErr{}})
}

func (s *SyncTestSuite) SetupTest() {
	testJobOne = nil
	testJobTwo = nil

	s.mockConfig.EXPECT().DefaultConnection().Return("sync").Once()
	s.mockConfig.EXPECT().DefaultQueue().Return("default").Once()
	s.mockConfig.EXPECT().Driver("sync").Return(queue.DriverSync).Once()
}

func (s *SyncTestSuite) TestDelay() {
	s.Nil(s.app.Job(&TestJobOne{}, testArgs).Delay(time.Now().Add(time.Second)).Dispatch())
	s.Equal(ConvertArgs(testArgs), testJobOne)
}

func (s *SyncTestSuite) TestDispatch() {
	s.Nil(s.app.Job(&TestJobOne{}, testArgs).Dispatch())
	s.Equal(ConvertArgs(testArgs), testJobOne)
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
	s.Nil(s.app.Chain([]queue.ChainJob{
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

	s.Equal(assert.AnError, s.app.Chain([]queue.ChainJob{
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

var (
	testArgs = []queue.Arg{
		{
			Type:  "bool",
			Value: true,
		},
		{
			Type:  "int",
			Value: 1,
		},
		{
			Type:  "int8",
			Value: int8(1),
		},
		{
			Type:  "int16",
			Value: int16(1),
		},
		{
			Type:  "int32",
			Value: int32(1),
		},
		{
			Type:  "int64",
			Value: int64(1),
		},
		{
			Type:  "uint",
			Value: uint(1),
		},
		{
			Type:  "uint8",
			Value: uint8(1),
		},
		{
			Type:  "uint16",
			Value: uint16(1),
		},
		{
			Type:  "uint32",
			Value: uint32(1),
		},
		{
			Type:  "uint64",
			Value: uint64(1),
		},
		{
			Type:  "float32",
			Value: float32(1.1),
		},
		{
			Type:  "float64",
			Value: float64(1.2),
		},
		{
			Type:  "string",
			Value: "test",
		},
		{
			Type:  "[]bool",
			Value: []bool{true, false},
		},
		{
			Type:  "[]int",
			Value: []int{1, 2, 3},
		},
		{
			Type:  "[]int8",
			Value: []int8{1, 2, 3},
		},
		{
			Type:  "[]int16",
			Value: []int16{1, 2, 3},
		},
		{
			Type:  "[]int32",
			Value: []int32{1, 2, 3},
		},
		{
			Type:  "[]int64",
			Value: []int64{1, 2, 3},
		},
		{
			Type:  "[]uint",
			Value: []uint{1, 2, 3},
		},
		{
			Type:  "[]uint8",
			Value: []uint8{1, 2, 3},
		},
		{
			Type:  "[]uint16",
			Value: []uint16{1, 2, 3},
		},
		{
			Type:  "[]uint32",
			Value: []uint32{1, 2, 3},
		},
		{
			Type:  "[]uint64",
			Value: []uint64{1, 2, 3},
		},
		{
			Type:  "[]float32",
			Value: []float32{1.1, 1.2, 1.3},
		},
		{
			Type:  "[]float64",
			Value: []float64{1.1, 1.2, 1.3},
		},
		{
			Type:  "[]string",
			Value: []string{"test", "test2", "test3"},
		},
	}
)
