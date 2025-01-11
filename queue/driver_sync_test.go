package queue

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/queue"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksqueue "github.com/goravel/framework/mocks/queue"
)

var (
	testSyncJob      = 0
	testChainSyncJob = 0
)

type DriverSyncTestSuite struct {
	suite.Suite
	app        *Application
	mockConfig *mocksconfig.Config
	mockQueue  *mocksqueue.Queue
}

func TestDriverSyncTestSuite(t *testing.T) {
	suite.Run(t, new(DriverSyncTestSuite))
}

func (s *DriverSyncTestSuite) SetupTest() {
	testSyncJob = 0
	testChainSyncJob = 0
	s.mockConfig = mocksconfig.NewConfig(s.T())
	s.mockQueue = mocksqueue.NewQueue(s.T())
	s.app = NewApplication(s.mockConfig)
	s.app.Register([]queue.Job{&TestSyncJob{}, &TestChainSyncJob{}})
	s.mockConfig.EXPECT().GetString("queue.default").Return("sync").Twice()
	s.mockConfig.EXPECT().GetString("queue.connections.sync.queue", "default").Return("default").Once()
}

func (s *DriverSyncTestSuite) TestSyncQueue() {
	s.mockConfig.EXPECT().GetString("app.name").Return("goravel").Once()
	s.Nil(s.app.Job(&TestSyncJob{}, []any{"TestSyncQueue", 1}).DispatchSync())
	s.Equal(1, testSyncJob)
}

func (s *DriverSyncTestSuite) TestChainSyncQueue() {
	s.mockConfig.EXPECT().GetString("app.name").Return("goravel").Twice()
	s.mockConfig.EXPECT().GetString("queue.connections.sync.driver").Return("sync").Once()

	s.Nil(s.app.Chain([]queue.Jobs{
		{
			Job:  &TestChainSyncJob{},
			Args: []any{"TestChainSyncJob", 1},
		},
		{
			Job:  &TestSyncJob{},
			Args: []any{"TestSyncJob", 1},
		},
	}).OnQueue("chain").Dispatch())

	time.Sleep(2 * time.Second)
	s.Equal(1, testChainSyncJob)
}

type TestSyncJob struct {
}

// Signature The name and signature of the job.
func (receiver *TestSyncJob) Signature() string {
	return "test_sync_job"
}

// Handle Execute the job.
func (receiver *TestSyncJob) Handle(args ...any) error {
	testSyncJob++

	return nil
}

type TestChainSyncJob struct {
}

// Signature The name and signature of the job.
func (receiver *TestChainSyncJob) Signature() string {
	return "test_chain_sync_job"
}

// Handle Execute the job.
func (receiver *TestChainSyncJob) Handle(args ...any) error {
	testChainSyncJob++

	return nil
}
