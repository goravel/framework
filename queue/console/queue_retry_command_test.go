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
	"github.com/goravel/framework/queue/models"
	"github.com/goravel/framework/queue/utils"
	"github.com/goravel/framework/support/carbon"
)

type QueueRetryCommandTestSuite struct {
	suite.Suite
	mockDB     *mocksdb.DB
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
				// getJobIDs
				var ids []string
				s.mockQuery.EXPECT().Pluck("id", &ids).Run(func(field string, dest any) {
					destPtr := dest.(*[]string)
					*destPtr = []string{"1"}
				}).Return(nil).Once()

				mockCtx.EXPECT().Info(errors.QueuePushingFailedJob.Error()).Once()
				mockCtx.EXPECT().Line("").Twice()

				var failedJob models.FailedJob
				s.mockQuery.EXPECT().Where("id", "1").Return(s.mockQuery).Once()
				s.mockQuery.EXPECT().FirstOrFail(&failedJob).Run(func(dest any) {
					destPtr := dest.(*models.FailedJob)
					*destPtr = models.FailedJob{
						ID:         1,
						UUID:       "test-uuid",
						Connection: "redis",
						Queue:      "default",
						Payload:    "{\"signature\":\"test_job\",\"args\":null,\"delay\":null,\"uuid\":\"test\",\"chain\":[]}",
					}
				}).Return(nil).Once()

				// retryJob
				mockJobStorer := mocksqueue.NewJobStorer(s.T())
				s.mockQueue.EXPECT().Connection("redis").Return(s.mockdriver, nil).Once()
				s.mockQueue.EXPECT().GetJobStorer().Return(mockJobStorer).Once()

				var task utils.Task
				s.mockJson.EXPECT().UnmarshalString("{\"signature\":\"test_job\",\"args\":null,\"delay\":null,\"uuid\":\"test\",\"chain\":[]}", &task).Run(func(payload string, taskPtr any) {
					taskPtr.(*utils.Task).Job.Signature = "test_job"
					taskPtr.(*utils.Task).UUID = "test"
				}).Return(nil).Once()
				mockJobStorer.EXPECT().Get("test_job").Return(&TestJob{}, nil).Once()
				s.mockdriver.EXPECT().Push(contractsqueue.Task{
					ChainJob: contractsqueue.ChainJob{
						Job: &TestJob{},
					},
					UUID: "test",
				}, "default").Return(nil).Once()

				s.mockQuery.EXPECT().Where("id", "1").Return(s.mockQuery).Once()
				s.mockQuery.EXPECT().Delete().Return(nil, nil).Once()

				mockCtx.EXPECT().TwoColumnDetail("test-uuid", "0s <fg=green;op=bold>DONE</>").Once()
			},
		},
		{
			name: "failed to get job IDs",
			setup: func() {
				var ids []string
				s.mockQuery.EXPECT().Pluck("id", &ids).Run(func(field string, dest any) {
					destPtr := dest.(*[]string)
					*destPtr = []string{"1"}
				}).Return(assert.AnError).Once()

				mockCtx.EXPECT().Error(assert.AnError.Error()).Once()
			},
		},
		{
			name: "no retryable jobs found",
			setup: func() {
				// getJobIDs
				var ids []string
				s.mockQuery.EXPECT().Pluck("id", &ids).Return(nil).Once()

				mockCtx.EXPECT().Info(errors.QueueNoRetryableJobsFound.Error()).Once()
			},
		},
		{
			name: "failed job not found",
			setup: func() {
				// getJobIDs
				var ids []string
				s.mockQuery.EXPECT().Pluck("id", &ids).Run(func(field string, dest any) {
					destPtr := dest.(*[]string)
					*destPtr = []string{"1"}
				}).Return(nil).Once()

				mockCtx.EXPECT().Info(errors.QueuePushingFailedJob.Error()).Once()
				mockCtx.EXPECT().Line("").Twice()

				var failedJob models.FailedJob
				s.mockQuery.EXPECT().Where("id", "1").Return(s.mockQuery).Once()
				s.mockQuery.EXPECT().FirstOrFail(&failedJob).Return(assert.AnError).Once()

				mockCtx.EXPECT().Error(assert.AnError.Error()).Once()
			},
		},
		{
			name: "failed to retry job",
			setup: func() {
				// getJobIDs
				var ids []string
				s.mockQuery.EXPECT().Pluck("id", &ids).Run(func(field string, dest any) {
					destPtr := dest.(*[]string)
					*destPtr = []string{"1"}
				}).Return(nil).Once()

				mockCtx.EXPECT().Info(errors.QueuePushingFailedJob.Error()).Once()
				mockCtx.EXPECT().Line("").Twice()

				var failedJob models.FailedJob
				destFailedJob := models.FailedJob{
					ID:         1,
					UUID:       "test-uuid",
					Connection: "redis",
					Queue:      "default",
					Payload:    "{\"signature\":\"test_job\",\"args\":null,\"delay\":null,\"uuid\":\"test\",\"chain\":[]}",
				}

				s.mockQuery.EXPECT().Where("id", "1").Return(s.mockQuery).Once()
				s.mockQuery.EXPECT().FirstOrFail(&failedJob).Run(func(dest any) {
					destPtr := dest.(*models.FailedJob)
					*destPtr = destFailedJob
				}).Return(nil).Once()

				// retryJob
				s.mockQueue.EXPECT().Connection("redis").Return(nil, assert.AnError).Once()

				mockCtx.EXPECT().Error(errors.QueueFailedToRetryJob.Args(destFailedJob, assert.AnError).Error()).Once()
			},
		},
		{
			name: "failed to delete failed job",
			setup: func() {
				// getJobIDs
				var ids []string
				s.mockQuery.EXPECT().Pluck("id", &ids).Run(func(field string, dest any) {
					destPtr := dest.(*[]string)
					*destPtr = []string{"1"}
				}).Return(nil).Once()

				mockCtx.EXPECT().Info(errors.QueuePushingFailedJob.Error()).Once()
				mockCtx.EXPECT().Line("").Twice()

				var failedJob models.FailedJob
				destFailedJob := models.FailedJob{
					ID:         1,
					UUID:       "test-uuid",
					Connection: "redis",
					Queue:      "default",
					Payload:    "{\"signature\":\"test_job\",\"args\":null,\"delay\":null,\"uuid\":\"test\",\"chain\":[]}",
				}
				s.mockQuery.EXPECT().Where("id", "1").Return(s.mockQuery).Once()
				s.mockQuery.EXPECT().FirstOrFail(&failedJob).Run(func(dest any) {
					destPtr := dest.(*models.FailedJob)
					*destPtr = destFailedJob
				}).Return(nil).Once()

				// retryJob
				mockJobStorer := mocksqueue.NewJobStorer(s.T())
				s.mockQueue.EXPECT().Connection("redis").Return(s.mockdriver, nil).Once()
				s.mockQueue.EXPECT().GetJobStorer().Return(mockJobStorer).Once()

				var task utils.Task
				s.mockJson.EXPECT().UnmarshalString("{\"signature\":\"test_job\",\"args\":null,\"delay\":null,\"uuid\":\"test\",\"chain\":[]}", &task).Run(func(payload string, taskPtr any) {
					taskPtr.(*utils.Task).Job.Signature = "test_job"
					taskPtr.(*utils.Task).UUID = "test"
				}).Return(nil).Once()
				mockJobStorer.EXPECT().Get("test_job").Return(&TestJob{}, nil).Once()
				s.mockdriver.EXPECT().Push(contractsqueue.Task{
					ChainJob: contractsqueue.ChainJob{
						Job: &TestJob{},
					},
					UUID: "test",
				}, "default").Return(nil).Once()

				s.mockQuery.EXPECT().Where("id", "1").Return(s.mockQuery).Once()
				s.mockQuery.EXPECT().Delete().Return(nil, assert.AnError).Once()

				mockCtx.EXPECT().Error(errors.QueueFailedToDeleteFailedJob.Args(destFailedJob, assert.AnError).Error()).Once()
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			mockCtx = mocksconsole.NewContext(s.T())
			mockCtx.EXPECT().Arguments().Return([]string{"all"}).Once()

			tt.setup()

			err := s.command.Handle(mockCtx)

			s.NoError(err)
		})
	}
}

func (s *QueueRetryCommandTestSuite) TestGetJobIDs() {
	tests := []struct {
		name        string
		args        []string
		options     map[string]string
		setup       func()
		expectedIDs []string
		expectError error
	}{
		{
			name: "get all jobs",
			args: []string{"all"},
			setup: func() {
				var ids []string
				s.mockQuery.EXPECT().Pluck("id", &ids).Run(func(field string, dest any) {
					destPtr := dest.(*[]string)
					*destPtr = []string{"1", "2", "3"}
				}).Return(nil).Once()
			},
			expectedIDs: []string{"1", "2", "3"},
		},
		{
			name: "get all jobs, database error",
			args: []string{"all"},
			setup: func() {
				var ids []string
				s.mockQuery.EXPECT().Pluck("id", &ids).Return(assert.AnError).Once()
			},
			expectError: assert.AnError,
		},
		{
			name: "get jobs by connection",
			args: []string{},
			options: map[string]string{
				"connection": "redis",
				"queue":      "",
			},
			setup: func() {
				s.mockQuery.EXPECT().Where("connection", "redis").Return(s.mockQuery).Once()

				var ids []string
				s.mockQuery.EXPECT().Pluck("id", &ids).Run(func(field string, dest any) {
					destPtr := dest.(*[]string)
					*destPtr = []string{"4", "5"}
				}).Return(nil).Once()
			},
			expectedIDs: []string{"4", "5"},
		},
		{
			name: "get jobs by connection, database error",
			args: []string{},
			options: map[string]string{
				"connection": "redis",
				"queue":      "",
			},
			setup: func() {
				s.mockQuery.EXPECT().Where("connection", "redis").Return(s.mockQuery).Once()

				var ids []string
				s.mockQuery.EXPECT().Pluck("id", &ids).Return(assert.AnError).Once()
			},
			expectError: assert.AnError,
		},
		{
			name: "get jobs by queue",
			args: []string{},
			options: map[string]string{
				"connection": "",
				"queue":      "default",
			},
			setup: func() {
				s.mockQuery.EXPECT().Where("queue", "default").Return(s.mockQuery).Once()

				var ids []string
				s.mockQuery.EXPECT().Pluck("id", &ids).Run(func(field string, dest any) {
					destPtr := dest.(*[]string)
					*destPtr = []string{"6", "7", "8"}
				}).Return(nil).Once()
			},
			expectedIDs: []string{"6", "7", "8"},
		},
		{
			name: "get jobs by connection and queue",
			args: []string{},
			options: map[string]string{
				"connection": "redis",
				"queue":      "default",
			},
			setup: func() {
				s.mockQuery.EXPECT().Where("connection", "redis").Return(s.mockQuery).Once()
				s.mockQuery.EXPECT().Where("queue", "default").Return(s.mockQuery).Once()

				var ids []string
				s.mockQuery.EXPECT().Pluck("id", &ids).Run(func(field string, dest any) {
					destPtr := dest.(*[]string)
					*destPtr = []string{"9"}
				}).Return(nil).Once()
			},
			expectedIDs: []string{"9"},
		},
		{
			name: "get jobs by UUIDs",
			args: []string{"uuid1", "uuid2"},
			options: map[string]string{
				"connection": "",
				"queue":      "",
			},
			setup: func() {
				s.mockQuery.EXPECT().WhereIn("uuid", []any{"uuid1", "uuid2"}).Return(s.mockQuery).Once()

				var ids []string
				s.mockQuery.EXPECT().Pluck("id", &ids).Run(func(field string, dest any) {
					destPtr := dest.(*[]string)
					*destPtr = []string{"10", "11"}
				}).Return(nil).Once()
			},
			expectedIDs: []string{"10", "11"},
		},
		{
			name: "get jobs by UUIDs, database error",
			args: []string{"uuid1", "uuid2"},
			options: map[string]string{
				"connection": "",
				"queue":      "",
			},
			setup: func() {
				s.mockQuery.EXPECT().WhereIn("uuid", []any{"uuid1", "uuid2"}).Return(s.mockQuery).Once()

				var ids []string
				s.mockQuery.EXPECT().Pluck("id", &ids).Return(assert.AnError).Once()
			},
			expectError: assert.AnError,
		},
		{
			name: "empty UUIDs",
			args: []string{},
			options: map[string]string{
				"connection": "",
				"queue":      "",
			},
			setup: func() {},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setup()

			mockCtx := mocksconsole.NewContext(s.T())
			mockCtx.EXPECT().Arguments().Return(tt.args).Once()
			if tt.options != nil {
				for key, value := range tt.options {
					mockCtx.EXPECT().Option(key).Return(value).Once()
				}
			}

			ids, err := s.command.getJobIDs(mockCtx)

			s.Equal(tt.expectedIDs, ids)
			s.Equal(tt.expectError, err)
		})
	}
}

func (s *QueueRetryCommandTestSuite) TestPrintSuccess() {
	mockCtx := mocksconsole.NewContext(s.T())
	mockCtx.EXPECT().TwoColumnDetail("test-uuid", "1s <fg=green;op=bold>DONE</>").Once()

	s.command.printSuccess(mockCtx, "test-uuid", "1s")
}

func (s *QueueRetryCommandTestSuite) TestRetryJob() {
	var jobStorer *mocksqueue.JobStorer

	failedJob := models.FailedJob{
		UUID:       "test-uuid",
		Connection: "redis",
		Queue:      "default",
		Payload:    "{\"signature\":\"test_job\",\"args\":null,\"delay\":null,\"uuid\":\"test\",\"chain\":[]}",
		Exception:  assert.AnError.Error(),
		FailedAt:   carbon.NewDateTime(carbon.Now()),
	}

	beforeEach := func() {
		jobStorer = mocksqueue.NewJobStorer(s.T())
	}

	tests := []struct {
		name        string
		failedJob   models.FailedJob
		setup       func()
		expectError error
	}{
		{
			name:      "successfully retry job",
			failedJob: failedJob,
			setup: func() {
				s.mockQueue.EXPECT().Connection("redis").Return(s.mockdriver, nil).Once()
				s.mockQueue.EXPECT().GetJobStorer().Return(jobStorer).Once()

				var task utils.Task
				s.mockJson.EXPECT().UnmarshalString(failedJob.Payload, &task).Run(func(payload string, taskPtr any) {
					taskPtr.(*utils.Task).Job.Signature = "test_job"
					taskPtr.(*utils.Task).UUID = "test"
				}).Return(nil).Once()
				jobStorer.EXPECT().Get("test_job").Return(&TestJob{}, nil).Once()
				s.mockdriver.EXPECT().Push(contractsqueue.Task{
					ChainJob: contractsqueue.ChainJob{
						Job: &TestJob{},
					},
					UUID: "test",
				}, failedJob.Queue).Return(nil).Once()
			},
		},
		{
			name:      "failed to get queue connection",
			failedJob: failedJob,
			setup: func() {
				s.mockQueue.EXPECT().Connection("redis").Return(nil, assert.AnError).Once()
			},
			expectError: assert.AnError,
		},
		{
			name: "error converting JSON to task",
			failedJob: models.FailedJob{
				ID:         1,
				UUID:       "test-uuid",
				Connection: "redis",
				Queue:      "default",
				Payload:    `invalid-json`,
			},
			setup: func() {
				s.mockQueue.EXPECT().Connection("redis").Return(s.mockdriver, nil).Once()
				s.mockQueue.EXPECT().GetJobStorer().Return(jobStorer).Once()

				var task utils.Task
				s.mockJson.EXPECT().UnmarshalString("invalid-json", &task).Return(assert.AnError).Once()
			},
			expectError: assert.AnError,
		},
		{
			name:      "error pushing task to queue",
			failedJob: failedJob,
			setup: func() {
				s.mockQueue.EXPECT().Connection("redis").Return(s.mockdriver, nil).Once()
				s.mockQueue.EXPECT().GetJobStorer().Return(jobStorer).Once()

				var task utils.Task
				s.mockJson.EXPECT().UnmarshalString(failedJob.Payload, &task).Run(func(payload string, taskPtr any) {
					taskPtr.(*utils.Task).Job.Signature = "test_job"
					taskPtr.(*utils.Task).UUID = "test"
				}).Return(nil).Once()
				jobStorer.EXPECT().Get("test_job").Return(&TestJob{}, nil).Once()
				s.mockdriver.EXPECT().Push(contractsqueue.Task{
					ChainJob: contractsqueue.ChainJob{
						Job: &TestJob{},
					},
					UUID: "test",
				}, failedJob.Queue).Return(assert.AnError).Once()
			},
			expectError: assert.AnError,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			beforeEach()
			tt.setup()

			err := s.command.retryJob(tt.failedJob)
			s.Equal(tt.expectError, err)
		})
	}
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
