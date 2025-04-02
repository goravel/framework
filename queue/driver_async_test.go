package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/queue"
	mocksqueue "github.com/goravel/framework/mocks/queue"
)

type DriverAsyncTestSuite struct {
	suite.Suite
	app        *Application
	mockConfig *mocksqueue.Config
}

func TestDriverAsyncTestSuite(t *testing.T) {
	suite.Run(t, new(DriverAsyncTestSuite))
}

func (s *DriverAsyncTestSuite) SetupTest() {
	testJobOne = nil
	testJobTwo = nil

	s.mockConfig = mocksqueue.NewConfig(s.T())

	s.mockConfig.EXPECT().DefaultConnection().Return("async").Once()
	s.mockConfig.EXPECT().Queue("async", "").Return("async_queue").Once()
	s.mockConfig.EXPECT().Driver("async").Return(queue.DriverAsync).Once()

	s.app = &Application{
		config: s.mockConfig,
		job:    NewJobRespository(),
	}

	s.app.Register([]queue.Job{&TestJobOne{}, &TestJobTwo{}, &TestJobErr{}})
}

func (s *DriverSyncTestSuite) TestDispatch() {
	args := []any{"a", 1, []string{"b", "c"}, []int{1, 2, 3}, map[string]any{"d": "e"}}

	s.Nil(s.app.Job(&TestJobOne{}, args).Dispatch())
	s.Equal(args, testJobOne)
}

func (s *DriverSyncTestSuite) TestChainDispatch() {
	argsOne := []any{"a", 1, []string{"b", "c"}, []int{1, 2, 3}, map[string]any{"d": "e"}}
	argsTwo := []any{"a", 2, []string{"d", "f"}, []int{4, 5, 6}, map[string]any{"g": "h"}}

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

	s.Equal(argsOne, testJobOne)
	s.Equal(argsTwo, testJobTwo)
}

func (s *DriverSyncTestSuite) TestChainDispatchWithError() {
	argsOne := []any{"a", 1, []string{"b", "c"}, []int{1, 2, 3}, map[string]any{"d": "e"}}
	argsTwo := []any{"a", 2, []string{"d", "f"}, []int{4, 5, 6}, map[string]any{"g": "h"}}

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

	s.Equal(argsOne, testJobOne)
	s.Nil(testJobTwo)
}
