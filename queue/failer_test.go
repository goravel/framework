package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database/db"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	mocksdb "github.com/goravel/framework/mocks/database/db"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mocksqueue "github.com/goravel/framework/mocks/queue"
	"github.com/goravel/framework/queue/models"
	"github.com/goravel/framework/queue/utils"
	"github.com/goravel/framework/support/carbon"
)

type FailerTestSuite struct {
	suite.Suite
	mockDB     *mocksdb.DB
	mockQuery  *mocksdb.Query
	mockQueue  *mocksqueue.Queue
	mockJson   *mocksfoundation.Json
	mockConfig *mocksqueue.Config
	failer     *Failer
}

func TestFailerTestSuite(t *testing.T) {
	suite.Run(t, new(FailerTestSuite))
}

func (s *FailerTestSuite) SetupTest() {
	s.mockDB = mocksdb.NewDB(s.T())
	s.mockQuery = mocksdb.NewQuery(s.T())
	s.mockQueue = mocksqueue.NewQueue(s.T())
	s.mockJson = mocksfoundation.NewJson(s.T())
	s.mockConfig = mocksqueue.NewConfig(s.T())

	s.mockConfig.EXPECT().GetString("queue.failed.database").Return("mysql").Once()
	s.mockConfig.EXPECT().GetString("queue.failed.table").Return("failed_jobs").Once()
	s.mockDB.EXPECT().Connection("mysql").Return(s.mockDB).Once()
	s.mockDB.EXPECT().Table("failed_jobs").Return(s.mockQuery).Once()

	s.failer = NewFailer(s.mockConfig, s.mockDB, s.mockQueue, s.mockJson)
}

func (s *FailerTestSuite) TestAll() {
	modelFailedJobs := []models.FailedJob{
		{
			UUID:       "test-uuid-1",
			Connection: "redis",
			Queue:      "default",
			Payload:    "test-payload-1",
			FailedAt:   carbon.NewDateTime(carbon.Now()),
		},
		{
			UUID:       "test-uuid-2",
			Connection: "redis",
			Queue:      "default",
			Payload:    "test-payload-2",
			FailedAt:   carbon.NewDateTime(carbon.Now()),
		},
	}

	tests := []struct {
		name          string
		setup         func()
		expectedJobs  []contractsqueue.FailedJob
		expectedError error
	}{
		{
			name: "success",
			setup: func() {
				var failedJobs []models.FailedJob
				s.mockQuery.EXPECT().Get(&failedJobs).Run(func(dest any) {
					*dest.(*[]models.FailedJob) = modelFailedJobs
				}).Return(nil).Once()
			},
			expectedJobs: []contractsqueue.FailedJob{
				NewFailedJob(modelFailedJobs[0], s.mockQuery, s.mockQueue, s.mockJson),
				NewFailedJob(modelFailedJobs[1], s.mockQuery, s.mockQueue, s.mockJson),
			},
		},
		{
			name: "error getting failed jobs",
			setup: func() {
				var failedJobs []models.FailedJob
				s.mockQuery.EXPECT().Get(&failedJobs).Return(assert.AnError).Once()
			},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setup()

			jobs, err := s.failer.All()

			if tt.expectedError != nil {
				s.Equal(tt.expectedError, err)
				s.Nil(jobs)
			} else {
				s.NoError(err)
				s.Len(jobs, len(tt.expectedJobs))
				for i, job := range jobs {
					s.Equal(tt.expectedJobs[i].UUID(), job.UUID())
					s.Equal(tt.expectedJobs[i].Connection(), job.Connection())
					s.Equal(tt.expectedJobs[i].Queue(), job.Queue())
				}
			}
		})
	}
}

func (s *FailerTestSuite) TestGet() {
	modelFailedJobs := []models.FailedJob{
		{
			UUID:       "test-uuid-1",
			Connection: "redis",
			Queue:      "default",
			Payload:    "test-payload-1",
			FailedAt:   carbon.NewDateTime(carbon.Now()),
		},
		{
			UUID:       "test-uuid-2",
			Connection: "redis",
			Queue:      "default",
			Payload:    "test-payload-2",
			FailedAt:   carbon.NewDateTime(carbon.Now()),
		},
	}

	tests := []struct {
		name          string
		connection    string
		queue         string
		uuids         []string
		setup         func()
		expectedJobs  []contractsqueue.FailedJob
		expectedError error
	}{
		{
			name:       "success with all filters",
			connection: "redis",
			queue:      "default",
			uuids:      []string{"test-uuid-1", "test-uuid-2"},
			setup: func() {
				s.mockQuery.EXPECT().Where("connection", "redis").Return(s.mockQuery).Once()
				s.mockQuery.EXPECT().Where("queue", "default").Return(s.mockQuery).Once()
				s.mockQuery.EXPECT().WhereIn("uuid", []any{"test-uuid-1", "test-uuid-2"}).Return(s.mockQuery).Once()

				var failedJobs []models.FailedJob
				s.mockQuery.EXPECT().Get(&failedJobs).Run(func(dest any) {
					*dest.(*[]models.FailedJob) = modelFailedJobs
				}).Return(nil).Once()
			},
			expectedJobs: []contractsqueue.FailedJob{
				NewFailedJob(modelFailedJobs[0], s.mockQuery, s.mockQueue, s.mockJson),
				NewFailedJob(modelFailedJobs[1], s.mockQuery, s.mockQueue, s.mockJson),
			},
		},
		{
			name:       "success with no filters",
			connection: "",
			queue:      "",
			uuids:      []string{},
			setup: func() {
				var failedJobs []models.FailedJob
				s.mockQuery.EXPECT().Get(&failedJobs).Run(func(dest any) {
					*dest.(*[]models.FailedJob) = modelFailedJobs
				}).Return(nil).Once()
			},
			expectedJobs: []contractsqueue.FailedJob{
				NewFailedJob(modelFailedJobs[0], s.mockQuery, s.mockQueue, s.mockJson),
				NewFailedJob(modelFailedJobs[1], s.mockQuery, s.mockQueue, s.mockJson),
			},
		},
		{
			name:       "error getting failed jobs",
			connection: "",
			queue:      "",
			uuids:      []string{},
			setup: func() {
				var failedJobs []models.FailedJob
				s.mockQuery.EXPECT().Get(&failedJobs).Return(assert.AnError).Once()
			},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setup()

			jobs, err := s.failer.Get(tt.connection, tt.queue, tt.uuids)

			if tt.expectedError != nil {
				s.Equal(tt.expectedError, err)
				s.Nil(jobs)
			} else {
				s.NoError(err)
				s.Len(jobs, len(tt.expectedJobs))
				for i, job := range jobs {
					s.Equal(tt.expectedJobs[i].UUID(), job.UUID())
					s.Equal(tt.expectedJobs[i].Connection(), job.Connection())
					s.Equal(tt.expectedJobs[i].Queue(), job.Queue())
				}
			}
		})
	}
}

type FailedJobTestSuite struct {
	suite.Suite
	mockDB         *mocksdb.DB
	mockQuery      *mocksdb.Query
	mockQueue      *mocksqueue.Queue
	mockJson       *mocksfoundation.Json
	modelFailedJob models.FailedJob
	failedJob      *FailedJob
}

func TestFailedJobTestSuite(t *testing.T) {
	suite.Run(t, new(FailedJobTestSuite))
}

func (s *FailedJobTestSuite) SetupTest() {
	s.mockDB = mocksdb.NewDB(s.T())
	s.mockQuery = mocksdb.NewQuery(s.T())
	s.mockQueue = mocksqueue.NewQueue(s.T())
	s.mockJson = mocksfoundation.NewJson(s.T())
	s.modelFailedJob = models.FailedJob{
		ID:         1,
		UUID:       "test-uuid",
		Connection: "redis",
		Queue:      "default",
		Payload:    `{"job":{"signature":"test-job"}, "uuid":"test-uuid"}`,
		FailedAt:   carbon.NewDateTime(carbon.Now()),
	}

	s.failedJob = NewFailedJob(s.modelFailedJob, s.mockQuery, s.mockQueue, s.mockJson)
}

func (s *FailedJobTestSuite) TestConnection() {
	s.Equal(s.modelFailedJob.Connection, s.failedJob.Connection())
}

func (s *FailedJobTestSuite) TestQueue() {
	s.Equal(s.modelFailedJob.Queue, s.failedJob.Queue())
}

func (s *FailedJobTestSuite) TestRetry() {
	tests := []struct {
		name          string
		setup         func()
		expectedError error
	}{
		{
			name: "happy path",
			setup: func() {
				mockDriver := mocksqueue.NewDriver(s.T())
				s.mockQueue.EXPECT().Connection(s.modelFailedJob.Connection).Return(mockDriver, nil).Once()

				mockJobStorer := mocksqueue.NewJobStorer(s.T())
				s.mockQueue.EXPECT().JobStorer().Return(mockJobStorer).Once()

				mockJob := mocksqueue.NewJob(s.T())
				mockJobStorer.EXPECT().Get("test-job").Return(mockJob, nil).Once()

				destTask := utils.Task{
					UUID: "test-uuid",
					Job: utils.Job{
						Signature: "test-job",
					},
				}

				var task utils.Task
				s.mockJson.EXPECT().UnmarshalString(s.modelFailedJob.Payload, &task).Run(func(json string, dest any) {
					*dest.(*utils.Task) = destTask
				}).Return(nil).Once()

				mockDriver.EXPECT().Push(contractsqueue.Task{
					UUID: "test-uuid",
					ChainJob: contractsqueue.ChainJob{
						Job: mockJob,
					},
				}, s.modelFailedJob.Queue).Return(nil).Once()
				s.mockQuery.EXPECT().Where("id", s.modelFailedJob.ID).Return(s.mockQuery).Once()
				s.mockQuery.EXPECT().Delete().Return(&db.Result{RowsAffected: 1}, nil).Once()
			},
		},
		{
			name: "failed to get connection",
			setup: func() {
				s.mockQueue.EXPECT().Connection(s.modelFailedJob.Connection).Return(nil, assert.AnError).Once()
			},
			expectedError: assert.AnError,
		},
		{
			name: "failed to decode payload",
			setup: func() {
				mockDriver := mocksqueue.NewDriver(s.T())
				s.mockQueue.EXPECT().Connection(s.modelFailedJob.Connection).Return(mockDriver, nil).Once()

				mockJobStorer := mocksqueue.NewJobStorer(s.T())
				s.mockQueue.EXPECT().JobStorer().Return(mockJobStorer).Once()

				var task utils.Task
				s.mockJson.EXPECT().UnmarshalString(s.modelFailedJob.Payload, &task).Return(assert.AnError).Once()
			},
			expectedError: assert.AnError,
		},
		{
			name: "failed to push task",
			setup: func() {
				mockDriver := mocksqueue.NewDriver(s.T())
				s.mockQueue.EXPECT().Connection(s.modelFailedJob.Connection).Return(mockDriver, nil).Once()

				mockJobStorer := mocksqueue.NewJobStorer(s.T())
				s.mockQueue.EXPECT().JobStorer().Return(mockJobStorer).Once()

				mockJob := mocksqueue.NewJob(s.T())
				mockJobStorer.EXPECT().Get("test-job").Return(mockJob, nil).Once()

				destTask := utils.Task{
					UUID: "test-uuid",
					Job: utils.Job{
						Signature: "test-job",
					},
				}

				var task utils.Task
				s.mockJson.EXPECT().UnmarshalString(s.modelFailedJob.Payload, &task).Run(func(json string, dest any) {
					*dest.(*utils.Task) = destTask
				}).Return(nil).Once()

				mockDriver.EXPECT().Push(contractsqueue.Task{
					UUID: "test-uuid",
					ChainJob: contractsqueue.ChainJob{
						Job: mockJob,
					},
				}, s.modelFailedJob.Queue).Return(assert.AnError).Once()
			},
			expectedError: assert.AnError,
		},
		{
			name: "failed to delete failed job",
			setup: func() {
				mockDriver := mocksqueue.NewDriver(s.T())
				s.mockQueue.EXPECT().Connection(s.modelFailedJob.Connection).Return(mockDriver, nil).Once()

				mockJobStorer := mocksqueue.NewJobStorer(s.T())
				s.mockQueue.EXPECT().JobStorer().Return(mockJobStorer).Once()

				mockJob := mocksqueue.NewJob(s.T())
				mockJobStorer.EXPECT().Get("test-job").Return(mockJob, nil).Once()

				destTask := utils.Task{
					UUID: "test-uuid",
					Job: utils.Job{
						Signature: "test-job",
					},
				}

				var task utils.Task
				s.mockJson.EXPECT().UnmarshalString(s.modelFailedJob.Payload, &task).Run(func(json string, dest any) {
					*dest.(*utils.Task) = destTask
				}).Return(nil).Once()

				mockDriver.EXPECT().Push(contractsqueue.Task{
					UUID: "test-uuid",
					ChainJob: contractsqueue.ChainJob{
						Job: mockJob,
					},
				}, s.modelFailedJob.Queue).Return(nil).Once()
				s.mockQuery.EXPECT().Where("id", s.modelFailedJob.ID).Return(s.mockQuery).Once()
				s.mockQuery.EXPECT().Delete().Return(nil, assert.AnError).Once()
			},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setup()

			err := s.failedJob.Retry()

			if tt.expectedError != nil {
				s.Equal(tt.expectedError, err)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *FailedJobTestSuite) TestFailedAt() {
	s.Equal(s.failedJob.failedJob.FailedAt, s.failedJob.FailedAt())
}

func (s *FailedJobTestSuite) TestSignature() {
	tests := []struct {
		name          string
		setup         func()
		expected      string
		expectedError error
	}{
		{
			name: "happy path",
			setup: func() {
				mockJobStorer := mocksqueue.NewJobStorer(s.T())
				s.mockQueue.EXPECT().JobStorer().Return(mockJobStorer).Once()

				mockJob := mocksqueue.NewJob(s.T())
				mockJobStorer.EXPECT().Get("test-job").Return(mockJob, nil).Once()
				mockJob.EXPECT().Signature().Return("test-signature").Once()

				destTask := utils.Task{
					UUID: "test-uuid",
					Job: utils.Job{
						Signature: "test-job",
					},
				}

				var task utils.Task
				s.mockJson.EXPECT().UnmarshalString(s.modelFailedJob.Payload, &task).Run(func(json string, dest any) {
					*dest.(*utils.Task) = destTask
				}).Return(nil).Once()
			},
			expected: "test-signature",
		},
		{
			name: "failed to decode payload",
			setup: func() {
				mockJobStorer := mocksqueue.NewJobStorer(s.T())
				s.mockQueue.EXPECT().JobStorer().Return(mockJobStorer).Once()

				var task utils.Task
				s.mockJson.EXPECT().UnmarshalString(s.modelFailedJob.Payload, &task).Return(assert.AnError).Once()
			},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setup()

			signature := s.failedJob.Signature()

			s.Equal(tt.expected, signature)
		})
	}
}

func (s *FailedJobTestSuite) TestUUID() {
	s.Equal("test-uuid", s.failedJob.UUID())
}
