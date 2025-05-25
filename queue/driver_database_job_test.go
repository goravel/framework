package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	mocksdb "github.com/goravel/framework/mocks/database/db"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mocksqueue "github.com/goravel/framework/mocks/queue"
	"github.com/goravel/framework/support/carbon"
)

type DatabaseReservedJobTestSuite struct {
	suite.Suite
	mockDB    *mocksdb.DB
	mockJob   *mocksqueue.JobStorer
	mockJson  *mocksfoundation.Json
	jobsTable string
}

func TestDatabaseReservedJobTestSuite(t *testing.T) {
	suite.Run(t, new(DatabaseReservedJobTestSuite))
}

func (s *DatabaseReservedJobTestSuite) SetupTest() {
	s.mockDB = mocksdb.NewDB(s.T())
	s.mockJob = mocksqueue.NewJobStorer(s.T())
	s.mockJson = mocksfoundation.NewJson(s.T())
	s.jobsTable = "jobs"
}

func (s *DatabaseReservedJobTestSuite) TestNewDatabaseReservedJob() {
	s.Run("happy path", func() {
		var task Task

		testJobOne := &TestJobOne{}
		s.mockJson.EXPECT().Unmarshal([]byte("{\"signature\":\"test_job_one\",\"args\":null,\"delay\":null,\"uuid\":\"test\",\"chain\":[]}"), &task).
			Run(func(_ []byte, taskPtr any) {
				taskPtr.(*Task).Job.Signature = testJobOne.Signature()
			}).Return(nil).Once()
		s.mockJob.EXPECT().Get(testJobOne.Signature()).Return(testJobOne, nil).Once()

		databaseJob := &DatabaseJob{
			ID:      1,
			Queue:   "default",
			Payload: "{\"signature\":\"test_job_one\",\"args\":null,\"delay\":null,\"uuid\":\"test\",\"chain\":[]}",
		}

		reservedJob, err := NewDatabaseReservedJob(databaseJob, s.mockDB, s.mockJob, s.mockJson, s.jobsTable)

		s.NoError(err)
		s.NotNil(reservedJob)
		s.Equal(databaseJob, reservedJob.job)
		s.Equal(s.mockDB, reservedJob.db)
		s.Equal(s.jobsTable, reservedJob.jobsTable)
	})

	s.Run("invalid payload", func() {
		var task Task

		s.mockJson.EXPECT().Unmarshal([]byte("invalid json"), &task).
			Return(assert.AnError).Once()

		databaseJob := &DatabaseJob{
			ID:      1,
			Queue:   "default",
			Payload: "invalid json",
		}

		reservedJob, err := NewDatabaseReservedJob(databaseJob, s.mockDB, s.mockJob, s.mockJson, s.jobsTable)

		s.Equal(assert.AnError, err)
		s.Nil(reservedJob)
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

			reservedJob := &DatabaseReservedJob{
				db:        s.mockDB,
				job:       &DatabaseJob{ID: id},
				jobsTable: s.jobsTable,
			}

			err := reservedJob.Delete()

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
			job := &DatabaseJob{
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

			job := &DatabaseJob{
				ReservedAt: test.initialReservedAt,
			}

			result := job.Touch()
			s.Equal(carbon.NewDateTime(now), result)
			s.Equal(carbon.NewDateTime(now), job.ReservedAt)
		})
	}
}
