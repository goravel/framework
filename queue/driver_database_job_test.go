package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	mocksdb "github.com/goravel/framework/mocks/database/db"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mocksqueue "github.com/goravel/framework/mocks/queue"
	"github.com/goravel/framework/queue/models"
	"github.com/goravel/framework/queue/utils"
	"github.com/goravel/framework/support/carbon"
)

type DatabaseReservedJobTestSuite struct {
	suite.Suite
	mockDB        *mocksdb.DB
	mockJobStorer *mocksqueue.JobStorer
	mockJson      *mocksfoundation.Json
	jobsTable     string
}

func TestDatabaseReservedJobTestSuite(t *testing.T) {
	suite.Run(t, new(DatabaseReservedJobTestSuite))
}

func (s *DatabaseReservedJobTestSuite) SetupTest() {
	s.mockDB = mocksdb.NewDB(s.T())
	s.mockJobStorer = mocksqueue.NewJobStorer(s.T())
	s.mockJson = mocksfoundation.NewJson(s.T())
	s.jobsTable = "jobs"
}

func (s *DatabaseReservedJobTestSuite) TestNewDatabaseReservedJob() {
	s.Run("happy path", func() {
		var task utils.Task

		testJobOne := &TestJobOne{}
		s.mockJson.EXPECT().UnmarshalString("{\"signature\":\"test_job_one\",\"args\":null,\"delay\":null,\"uuid\":\"test\",\"chain\":[]}", &task).
			Run(func(_ string, taskPtr any) {
				taskPtr.(*utils.Task).Job.Signature = testJobOne.Signature()
			}).Return(nil).Once()
		s.mockJobStorer.EXPECT().Get(testJobOne.Signature()).Return(testJobOne, nil).Once()

		job := &models.Job{
			ID:      1,
			Queue:   "default",
			Payload: "{\"signature\":\"test_job_one\",\"args\":null,\"delay\":null,\"uuid\":\"test\",\"chain\":[]}",
		}

		databaseReservedJob, err := NewDatabaseReservedJob(job, s.mockDB, s.mockJobStorer, s.mockJson, s.jobsTable)

		s.NoError(err)
		s.NotNil(databaseReservedJob)
		s.Equal(job, databaseReservedJob.job)
		s.Equal(s.mockDB, databaseReservedJob.db)
		s.Equal(s.jobsTable, databaseReservedJob.jobsTable)
	})

	s.Run("invalid payload", func() {
		var task utils.Task

		s.mockJson.EXPECT().UnmarshalString("invalid json", &task).
			Return(assert.AnError).Once()

		job := &models.Job{
			ID:      1,
			Queue:   "default",
			Payload: "invalid json",
		}

		databaseReservedJob, err := NewDatabaseReservedJob(job, s.mockDB, s.mockJobStorer, s.mockJson, s.jobsTable)

		s.Equal(assert.AnError, err)
		s.Nil(databaseReservedJob)
	})
}

func (s *DatabaseReservedJobTestSuite) TestDelete() {
	tests := []struct {
		name          string
		expectedError error
	}{
		{
			name:          "happy path",
			expectedError: nil,
		},
		{
			name:          "error",
			expectedError: assert.AnError,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			id := uint(1)
			mockQuery := mocksdb.NewQuery(s.T())
			s.mockDB.EXPECT().Table(s.jobsTable).Return(mockQuery).Once()
			mockQuery.EXPECT().Where("id", id).Return(mockQuery).Once()
			mockQuery.EXPECT().Delete().Return(nil, test.expectedError).Once()

			databsaeReservedJob := &DatabaseReservedJob{
				db:        s.mockDB,
				job:       &models.Job{ID: id},
				jobsTable: s.jobsTable,
			}

			err := databsaeReservedJob.Delete()

			s.Equal(test.expectedError, err)
		})
	}
}

func (s *DatabaseReservedJobTestSuite) TestTask() {
	task := contractsqueue.Task{
		ChainJob: contractsqueue.ChainJob{
			Job: &TestJobOne{},
		},
	}

	reservedJob := &DatabaseReservedJob{
		task: task,
	}

	result := reservedJob.Task()
	s.Equal(task, result)
}

type DatabaseJobTestSuite struct {
	suite.Suite
}

func TestDatabaseJobTestSuite(t *testing.T) {
	suite.Run(t, new(DatabaseJobTestSuite))
}

func (s *DatabaseJobTestSuite) TestIncrement() {
	tests := []struct {
		name            string
		initialAttempts int
		expectedResult  int
	}{
		{
			name:            "increment from zero",
			initialAttempts: 0,
			expectedResult:  1,
		},
		{
			name:            "increment from positive number",
			initialAttempts: 1,
			expectedResult:  2,
		},
		{
			name:            "increment from negative number",
			initialAttempts: -1,
			expectedResult:  0,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			job := &models.Job{
				Attempts: test.initialAttempts,
			}

			result := job.Increment()
			s.Equal(test.expectedResult, result)
			s.Equal(test.expectedResult, job.Attempts)
		})
	}
}

func (s *DatabaseJobTestSuite) TestTouch() {
	tests := []struct {
		name              string
		initialReservedAt *carbon.DateTime
	}{
		{
			name:              "touch with nil reserved_at",
			initialReservedAt: nil,
		},
		{
			name:              "touch with existing reserved_at",
			initialReservedAt: carbon.NewDateTime(carbon.Now().SubHour()),
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			now := carbon.Now()
			carbon.SetTestNow(now)

			job := &models.Job{
				ReservedAt: test.initialReservedAt,
			}

			result := job.Touch()
			s.Equal(carbon.NewDateTime(now), result)
			s.Equal(carbon.NewDateTime(now), job.ReservedAt)
		})
	}
}
