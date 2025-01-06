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
	mockConfig := mocksconfig.NewConfig(t)
	mockQueue := mocksqueue.NewQueue(t)
	app := NewApplication(mockConfig)

	app.Register([]queue.Job{&TestSyncJob{}, &TestChainSyncJob{}})
	suite.Run(t, &DriverSyncTestSuite{
		app:        app,
		mockConfig: mockConfig,
		mockQueue:  mockQueue,
	})
}

func (s *DriverSyncTestSuite) SetupTest() {
	s.mockConfig.On("GetString", "queue.default").Return("sync").Once()
	s.mockConfig.On("GetString", "app.name").Return("goravel").Once()
	s.mockConfig.On("GetString", "queue.connections.sync.queue", "default").Return("default").Once()
	testSyncJob = 0
	testChainSyncJob = 0
}

func (s *DriverSyncTestSuite) TestSyncQueue() {
	s.Nil(s.app.Job(&TestSyncJob{}, []any{"TestSyncQueue", 1}).DispatchSync())
	s.Equal(1, testSyncJob)
}

func (s *DriverSyncTestSuite) TestChainSyncQueue() {
	s.mockConfig.On("GetString", "queue.default").Return("sync").Times(2)
	s.mockConfig.On("GetString", "app.name").Return("goravel").Once()
	s.mockConfig.On("GetString", "queue.connections.sync.driver").Return("sync").Once()

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
