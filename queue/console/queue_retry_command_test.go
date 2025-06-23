package console

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksdb "github.com/goravel/framework/mocks/database/db"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mocksqueue "github.com/goravel/framework/mocks/queue"
	"github.com/goravel/framework/support/carbon"
)

type QueueRetryCommandTestSuite struct {
	suite.Suite
	mockDB     *mocksdb.DB
	mockFailer *mocksqueue.Failer
	mockQuery  *mocksdb.Query
	mockQueue  *mocksqueue.Queue
	mockJson   *mocksfoundation.Json
	mockConfig *mocksqueue.Config
	mockdriver *mocksqueue.Driver
	command    *QueueRetryCommand
}

func TestQueueRetryCommandTestSuite(t *testing.T) {
	suite.Run(t, new(QueueRetryCommandTestSuite))
}

func (s *QueueRetryCommandTestSuite) SetupTest() {
	s.mockDB = mocksdb.NewDB(s.T())
	s.mockQuery = mocksdb.NewQuery(s.T())
	s.mockFailer = mocksqueue.NewFailer(s.T())
	s.mockQueue = mocksqueue.NewQueue(s.T())
	s.mockJson = mocksfoundation.NewJson(s.T())
	s.mockConfig = mocksqueue.NewConfig(s.T())
	s.mockdriver = mocksqueue.NewDriver(s.T())

	s.mockConfig.EXPECT().GetString("queue.failed.database").Return("mysql").Once()
	s.mockConfig.EXPECT().GetString("queue.failed.table").Return("failed_jobs").Once()
	s.mockDB.EXPECT().Connection("mysql").Return(s.mockDB).Once()
	s.mockDB.EXPECT().Table("failed_jobs").Return(s.mockQuery).Once()

	s.command = NewQueueRetryCommand(s.mockConfig, s.mockDB, s.mockQueue, s.mockJson)
}

func (s *QueueRetryCommandTestSuite) TestHandle() {
	var mockCtx *mocksconsole.Context

	carbon.SetTestNow(carbon.Now())
	defer carbon.ClearTestNow()

	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "happy path",
			setup: func() {
				mockCtx.EXPECT().Arguments().Return([]string{"1"}).Once()
				mockCtx.EXPECT().Option("connection").Return("redis").Once()
				mockCtx.EXPECT().Option("queue").Return("default").Once()
				s.mockQueue.EXPECT().Failer().Return(s.mockFailer).Once()

				mockFailedJob := mocksqueue.NewFailedJob(s.T())
				s.mockFailer.EXPECT().Get("redis", "default", []string{"1"}).Return([]contractsqueue.FailedJob{
					mockFailedJob,
				}, nil).Once()

				mockCtx.EXPECT().Info(errors.QueuePushingFailedJob.Error()).Once()
				mockCtx.EXPECT().Line("").Twice()
				mockFailedJob.EXPECT().Retry().Return(nil).Once()
				mockFailedJob.EXPECT().UUID().Return("test-uuid").Once()
				mockCtx.EXPECT().TwoColumnDetail("test-uuid", "0s <fg=green;op=bold>DONE</>").Once()
			},
		},
		{
			name: "one failed job retrys failed, another one success",
			setup: func() {
				mockCtx.EXPECT().Arguments().Return([]string{"1"}).Once()
				mockCtx.EXPECT().Option("connection").Return("redis").Once()
				mockCtx.EXPECT().Option("queue").Return("default").Once()
				s.mockQueue.EXPECT().Failer().Return(s.mockFailer).Once()

				mockFailedJobOne := mocksqueue.NewFailedJob(s.T())
				mockFailedJobTwo := mocksqueue.NewFailedJob(s.T())
				s.mockFailer.EXPECT().Get("redis", "default", []string{"1"}).Return([]contractsqueue.FailedJob{
					mockFailedJobOne,
					mockFailedJobTwo,
				}, nil).Once()

				mockCtx.EXPECT().Info(errors.QueuePushingFailedJob.Error()).Once()
				mockCtx.EXPECT().Line("").Twice()
				mockFailedJobOne.EXPECT().Retry().Return(assert.AnError).Once()
				mockCtx.EXPECT().Error(errors.QueueFailedToRetryJob.Args(mockFailedJobOne, assert.AnError).Error()).Once()
				mockFailedJobTwo.EXPECT().Retry().Return(nil).Once()
				mockFailedJobTwo.EXPECT().UUID().Return("test-uuid-2").Once()
				mockCtx.EXPECT().TwoColumnDetail("test-uuid-2", "0s <fg=green;op=bold>DONE</>").Once()
			},
		},
		{
			name: "failed job query is nil",
			setup: func() {
				s.command.failedJobQuery = nil

				mockCtx.EXPECT().Error(errors.DBFacadeNotSet.Error()).Once()
			},
		},
		{
			name: "failed to get failed jobs",
			setup: func() {
				mockCtx.EXPECT().Arguments().Return([]string{"1"}).Once()
				mockCtx.EXPECT().Option("connection").Return("redis").Once()
				mockCtx.EXPECT().Option("queue").Return("default").Once()
				s.mockQueue.EXPECT().Failer().Return(s.mockFailer).Once()
				s.mockFailer.EXPECT().Get("redis", "default", []string{"1"}).Return(nil, assert.AnError).Once()
				mockCtx.EXPECT().Error(assert.AnError.Error()).Once()
			},
		},
		{
			name: "no retryable jobs found",
			setup: func() {
				mockCtx.EXPECT().Arguments().Return([]string{"1"}).Once()
				mockCtx.EXPECT().Option("connection").Return("redis").Once()
				mockCtx.EXPECT().Option("queue").Return("default").Once()
				s.mockQueue.EXPECT().Failer().Return(s.mockFailer).Once()
				s.mockFailer.EXPECT().Get("redis", "default", []string{"1"}).Return(nil, nil).Once()
				mockCtx.EXPECT().Info(errors.QueueNoRetryableJobsFound.Error()).Once()
			},
		},
		{
			name: "failed to retry job",
			setup: func() {
				mockCtx.EXPECT().Arguments().Return([]string{"1"}).Once()
				mockCtx.EXPECT().Option("connection").Return("redis").Once()
				mockCtx.EXPECT().Option("queue").Return("default").Once()
				s.mockQueue.EXPECT().Failer().Return(s.mockFailer).Once()

				mockFailedJob := mocksqueue.NewFailedJob(s.T())
				s.mockFailer.EXPECT().Get("redis", "default", []string{"1"}).Return([]contractsqueue.FailedJob{
					mockFailedJob,
				}, nil).Once()

				mockCtx.EXPECT().Info(errors.QueuePushingFailedJob.Error()).Once()
				mockCtx.EXPECT().Line("").Twice()
				mockFailedJob.EXPECT().Retry().Return(assert.AnError).Once()
				mockCtx.EXPECT().Error(errors.QueueFailedToRetryJob.Args(mockFailedJob, assert.AnError).Error()).Once()
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()

			mockCtx = mocksconsole.NewContext(s.T())

			tt.setup()

			err := s.command.Handle(mockCtx)

			s.NoError(err)
		})
	}
}

func (s *QueueRetryCommandTestSuite) TestPrintSuccess() {
	mockCtx := mocksconsole.NewContext(s.T())
	mockCtx.EXPECT().TwoColumnDetail("test-uuid", "1s <fg=green;op=bold>DONE</>").Once()

	s.command.printSuccess(mockCtx, "test-uuid", "1s")
}

type TestJob struct {
}

// Signature The name and signature of the job.
func (r *TestJob) Signature() string {
	return "test_job"
}

// Handle Execute the job.
func (r *TestJob) Handle(args ...any) error {
	return nil
}
