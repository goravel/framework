package queue

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
)

type WorkerTestSuite struct {
	suite.Suite
	app        *Application
	mockConfig *mocksconfig.Config
}

func TestWorkerTestSuite(t *testing.T) {
	suite.Run(t, new(WorkerTestSuite))
}

func (s *WorkerTestSuite) SetupTest() {
	s.mockConfig = mocksconfig.NewConfig(s.T())
	s.mockConfig.EXPECT().GetString("queue.default").Return("async").Times(3)
	s.mockConfig.EXPECT().GetString("app.name").Return("goravel").Times(2)
	s.mockConfig.EXPECT().GetString("queue.connections.async.queue", "default").Return("default").Times(2)
	s.mockConfig.EXPECT().GetString("queue.connections.async.driver").Return("async").Twice()
	s.mockConfig.EXPECT().GetInt("queue.connections.async.size", 100).Return(10).Twice()
	s.app = NewApplication(s.mockConfig)
}

func (s *WorkerTestSuite) TestRun_Success() {
	testJob := new(MockJob)
	s.app.Register([]contractsqueue.Job{testJob})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	go func(ctx context.Context) {
		s.NoError(s.app.Worker().Run())
		<-ctx.Done()
		s.NoError(s.app.Worker().Shutdown())
	}(ctx)

	time.Sleep(1 * time.Second)
	s.NoError(s.app.Job(testJob, []any{}).Dispatch())
	time.Sleep(2 * time.Second)
	s.True(testJob.called)
	cancel()
	time.Sleep(2 * time.Second)
}

func (s *WorkerTestSuite) TestRun_FailedJob() {
	s.mockConfig.EXPECT().GetString("queue.failed.database").Return("database").Times(1)
	s.mockConfig.EXPECT().GetString("queue.failed.table").Return("failed_jobs").Times(1)

	mockOrm := mocksorm.NewOrm(s.T())
	mockQuery := mocksorm.NewQuery(s.T())
	mockOrm.EXPECT().Connection("database").Return(mockOrm)
	mockOrm.EXPECT().Query().Return(mockQuery)
	mockQuery.EXPECT().Table("failed_jobs").Return(mockQuery)
	mockQuery.EXPECT().Create(mock.Anything).Return(nil)
	OrmFacade = mockOrm

	testJob := new(MockFailedJob)
	s.app.Register([]contractsqueue.Job{testJob})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	go func(ctx context.Context) {
		s.NoError(s.app.Worker().Run())
		<-ctx.Done()
		s.NoError(s.app.Worker().Shutdown())
	}(ctx)

	time.Sleep(1 * time.Second)
	s.NoError(s.app.Job(testJob, []any{}).Dispatch())
	time.Sleep(2 * time.Second)
	s.True(testJob.called)
	cancel()
	time.Sleep(2 * time.Second)
}

type MockFailedJob struct {
	signature string
	called    bool
}

func (m *MockFailedJob) Signature() string {
	return m.signature
}

func (m *MockFailedJob) Handle(args ...any) error {
	m.called = true
	return errors.New("failed job")
}
