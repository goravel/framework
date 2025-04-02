package queue

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
	mocksqueue "github.com/goravel/framework/mocks/queue"
	"github.com/goravel/framework/support/carbon"
)

type DriverAsyncTestSuite struct {
	suite.Suite
	app        *Application
	mockConfig *mocksqueue.Config
}

func TestDriverAsyncTestSuite(t *testing.T) {
	suite.Run(t, new(DriverAsyncTestSuite))
}

func (s *DriverAsyncTestSuite) SetupSuite() {
	s.mockConfig = mocksqueue.NewConfig(s.T())

	s.app = &Application{
		config: s.mockConfig,
		job:    NewJobRespository(),
	}

	s.app.Register([]contractsqueue.Job{&TestJobOne{}, &TestJobTwo{}, &TestJobErr{}})

	go func() {
		s.mockConfig.EXPECT().DefaultConnection().Return("async").Once()
		s.mockConfig.EXPECT().Queue("async", "").Return("default_queue").Once()
		s.mockConfig.EXPECT().Driver("async").Return(contractsqueue.DriverAsync).Once()
		s.mockConfig.EXPECT().Size("async").Return(1).Once()

		s.NoError(s.app.Worker().Run())
	}()

	go func() {
		s.mockConfig.EXPECT().DefaultConnection().Return("async").Once()
		s.mockConfig.EXPECT().Queue("async", "another").Return("another_queue").Once()
		s.mockConfig.EXPECT().Driver("async").Return(contractsqueue.DriverAsync).Once()
		s.mockConfig.EXPECT().Size("async").Return(1).Once()

		s.NoError(s.app.Worker(contractsqueue.Args{
			Connection: "async",
			Queue:      "another",
			Concurrent: 1,
		}).Run())
	}()

	time.Sleep(1 * time.Second)
}

func (s *DriverAsyncTestSuite) SetupTest() {
	testJobOne = nil
	testJobTwo = nil

	s.mockConfig.EXPECT().DefaultConnection().Return("async").Once()
	s.mockConfig.EXPECT().Queue("async", "").Return("default_queue").Once()
	s.mockConfig.EXPECT().Driver("async").Return(contractsqueue.DriverAsync).Once()
	s.mockConfig.EXPECT().Size("async").Return(1).Once()
}

func (s *DriverAsyncTestSuite) TestDispatch() {
	args := []any{"a", 1, []string{"b", "c"}, []int{1, 2, 3}, map[string]any{"d": "e"}}

	s.Nil(s.app.Job(&TestJobOne{}, args).Dispatch())

	time.Sleep(2 * time.Second)

	s.Equal(args, testJobOne)
}

func (s *DriverAsyncTestSuite) TestChainDispatch() {
	argsOne := []any{"a", 1, []string{"b", "c"}, []int{1, 2, 3}, map[string]any{"d": "e"}}
	argsTwo := []any{"a", 2, []string{"d", "f"}, []int{4, 5, 6}, map[string]any{"g": "h"}}

	s.Nil(s.app.Chain([]contractsqueue.Jobs{
		{
			Job:  &TestJobOne{},
			Args: argsOne,
		},
		{
			Job:  &TestJobTwo{},
			Args: argsTwo,
		},
	}).Dispatch())

	time.Sleep(2 * time.Second)

	s.Equal(argsOne, testJobOne)
	s.Equal(argsTwo, testJobTwo)
}

func (s *DriverAsyncTestSuite) TestChainDispatchWithError() {
	now := carbon.Now()
	carbon.SetTestNow(now)
	defer carbon.UnsetTestNow()

	argsOne := []any{"a", 1, []string{"b", "c"}, []int{1, 2, 3}, map[string]any{"d": "e"}}
	argsTwo := []any{"a", 2, []string{"d", "f"}, []int{4, 5, 6}, map[string]any{"g": "h"}}
	argsErr := []any{"a", 3, []string{"d", "f"}, []int{4, 5, 6}, map[string]any{"g": "h"}}

	mockQuery := mocksorm.NewQuery(s.T())
	s.mockConfig.EXPECT().FailedJobsQuery().Return(mockQuery).Once()
	mockQuery.EXPECT().Create(mock.MatchedBy(func(job *FailedJob) bool {
		return len(job.UUID) > 0 &&
			job.Connection == "async" &&
			job.Queue == "default_queue" &&
			s.ElementsMatch(job.Payload, argsErr) &&
			job.Exception == assert.AnError.Error() &&
			job.FailedAt.Eq(now)
	})).Return(nil).Once()

	s.NoError(s.app.Chain([]contractsqueue.Jobs{
		{
			Job:  &TestJobOne{},
			Args: argsOne,
		},
		{
			Job:  &TestJobErr{},
			Args: argsErr,
		},
		{
			Job:  &TestJobTwo{},
			Args: argsTwo,
		},
	}).Dispatch())

	time.Sleep(2 * time.Second)

	s.Equal(argsOne, testJobOne)
	s.Nil(testJobTwo)
}
