package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	contractsdb "github.com/goravel/framework/contracts/database/db"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	mocksdb "github.com/goravel/framework/mocks/database/db"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mocksqueue "github.com/goravel/framework/mocks/queue"
	"github.com/goravel/framework/queue/models"
	"github.com/goravel/framework/queue/utils"
	"github.com/goravel/framework/support/carbon"
)

type DatabaseTestSuite struct {
	suite.Suite
	mockDB        *mocksdb.DB
	mockJobStorer *mocksqueue.JobStorer
	mockJson      *mocksfoundation.Json
	jobsTable     string
	retryAfter    int
	connection    string
	database      *Database
}

func TestDatabaseTestSuite(t *testing.T) {
	suite.Run(t, new(DatabaseTestSuite))
}

func (s *DatabaseTestSuite) SetupTest() {
	s.mockDB = mocksdb.NewDB(s.T())
	s.mockJobStorer = mocksqueue.NewJobStorer(s.T())
	s.mockJson = mocksfoundation.NewJson(s.T())
	s.jobsTable = "jobs"
	s.retryAfter = 60
	s.connection = "default"
	s.database = &Database{
		db:         s.mockDB,
		jobStorer:  s.mockJobStorer,
		json:       s.mockJson,
		jobsTable:  s.jobsTable,
		retryAfter: s.retryAfter,
	}
}

func (s *DatabaseTestSuite) TestNewDatabase() {
	var mockConfig *mocksqueue.Config

	tests := []struct {
		name          string
		setup         func()
		expectedError error
	}{
		{
			name: "successful creation",
			setup: func() {
				mockConfig.EXPECT().GetString("queue.connections.default.connection").Return("mysql").Once()
				mockConfig.EXPECT().GetString("queue.connections.default.table", "jobs").Return("jobs").Once()
				mockConfig.EXPECT().GetInt("queue.connections.default.retry_after", 60).Return(60).Once()
				s.mockDB.EXPECT().Connection("mysql").Return(s.mockDB).Once()
			},
			expectedError: nil,
		},
		{
			name: "invalid database connection",
			setup: func() {
				mockConfig.EXPECT().GetString("queue.connections.default.connection").Return("").Once()
			},
			expectedError: errors.QueueInvalidDatabaseConnection.Args(""),
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			mockConfig = mocksqueue.NewConfig(s.T())

			test.setup()

			database, err := NewDatabase(mockConfig, s.mockDB, s.mockJobStorer, s.mockJson, s.connection)

			if test.expectedError != nil {
				s.Equal(test.expectedError, err)
				s.Nil(database)
			} else {
				s.NoError(err)
				s.NotNil(database)
				s.Equal(s.mockDB, database.db)
				s.Equal(s.mockJobStorer, database.jobStorer)
				s.Equal(s.mockJson, database.json)
				s.Equal(s.jobsTable, database.jobsTable)
				s.Equal(s.retryAfter, database.retryAfter)
			}
		})
	}
}

func (s *DatabaseTestSuite) TestDriver() {
	s.Equal(contractsqueue.DriverDatabase, s.database.Driver())
}

func (s *DatabaseTestSuite) TestPop() {
	queue := "default"
	payload := "{\"signature\":\"test_job_one\",\"args\":null,\"delay\":null,\"uuid\":\"test\",\"chain\":[]}"
	testJobOne := &TestJobOne{}

	carbon.SetTestNow(carbon.Now())
	defer carbon.ClearTestNow()

	tests := []struct {
		name            string
		setup           func()
		wantReservedJob contractsqueue.ReservedJob
		wantError       error
	}{
		{
			name: "happy path",
			setup: func() {
				mockTx := mocksdb.NewTx(s.T())
				mockQuery := mocksdb.NewQuery(s.T())

				s.mockDB.EXPECT().Transaction(mock.Anything).Run(func(txFunc func(tx contractsdb.Tx) error) {
					s.NoError(txFunc(mockTx))
				}).Return(nil).Once()

				mockTx.EXPECT().Table(s.jobsTable).Return(mockQuery).Once()
				mockQuery.EXPECT().LockForUpdate().Return(mockQuery).Once()
				mockQuery.EXPECT().Where("queue", queue).Return(mockQuery).Once()
				mockQuery.EXPECT().Where(mock.Anything).Return(mockQuery).Once()
				mockQuery.EXPECT().OrderBy("id").Return(mockQuery).Once()

				var job models.Job
				wantJob := models.Job{
					ID:       1,
					Queue:    queue,
					Payload:  payload,
					Attempts: 0,
				}
				mockQuery.EXPECT().First(&job).
					Run(func(dest any) {
						*dest.(*models.Job) = wantJob
					}).Return(nil).Once()

				mockTx.EXPECT().Table(s.jobsTable).Return(mockQuery).Once()
				mockQuery.EXPECT().Where("id", uint(1)).Return(mockQuery).Once()
				mockQuery.EXPECT().Update(map[string]any{
					"attempts":    1,
					"reserved_at": carbon.NewDateTime(carbon.Now()),
				}).Return(nil, nil).Once()

				var task utils.Task
				s.mockJson.EXPECT().UnmarshalString(payload, &task).
					Run(func(_ string, taskPtr any) {
						*taskPtr.(*utils.Task) = utils.Task{
							UUID: "test",
							Job: utils.Job{
								Signature: testJobOne.Signature(),
							},
						}
					}).Return(nil).Once()
				s.mockJobStorer.EXPECT().Get(testJobOne.Signature()).Return(testJobOne, nil).Once()
			},
			wantReservedJob: &DatabaseReservedJob{
				db: s.mockDB,
				job: &models.Job{
					ID:         1,
					Queue:      queue,
					Payload:    payload,
					Attempts:   1,
					ReservedAt: carbon.NewDateTime(carbon.Now()),
				},
				jobsTable: s.jobsTable,
				task: contractsqueue.Task{
					UUID: "test",
					ChainJob: contractsqueue.ChainJob{
						Job: testJobOne,
					},
				},
			},
			wantError: nil,
		},
		{
			name: "no job found",
			setup: func() {
				s.mockDB.EXPECT().Transaction(mock.Anything).Return(errors.QueueDriverNoJobFound.Args(queue)).Once()
			},
			wantError: errors.QueueDriverNoJobFound.Args(queue),
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			test.setup()

			job, err := s.database.Pop(queue)

			s.Equal(test.wantReservedJob, job)
			s.Equal(test.wantError, err)
		})
	}
}

func (s *DatabaseTestSuite) TestPush() {
	testJobOne := &TestJobOne{}
	payload := "{\"signature\":\"test_job_one\",\"args\":null,\"delay\":null,\"uuid\":\"test\",\"chain\":[]}"
	queue := "default"
	task := contractsqueue.Task{
		UUID: "test",
		ChainJob: contractsqueue.ChainJob{
			Job: &TestJobOne{},
		},
	}
	internalTask := utils.Task{
		UUID: "test",
		Job: utils.Job{
			Signature: testJobOne.Signature(),
		},
		Chain: []utils.Job{},
	}
	carbon.SetTestNow(carbon.Now())
	defer carbon.ClearTestNow()

	tests := []struct {
		name          string
		setup         func()
		expectedError error
	}{
		{
			name: "successful push",
			setup: func() {
				s.mockJson.EXPECT().MarshalString(internalTask).Return(payload, nil).Once()
				mockQuery := mocksdb.NewQuery(s.T())
				s.mockDB.EXPECT().Table(s.jobsTable).Return(mockQuery).Once()
				mockQuery.EXPECT().Insert(&models.Job{
					Queue:       queue,
					Payload:     payload,
					AvailableAt: carbon.NewDateTime(carbon.Now()),
					CreatedAt:   carbon.NewDateTime(carbon.Now()),
				}).Return(&contractsdb.Result{RowsAffected: 1}, nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "failed to marshal task",
			setup: func() {
				s.mockJson.EXPECT().MarshalString(internalTask).Return("", assert.AnError).Once()
			},
			expectedError: assert.AnError,
		},
		{
			name: "failed to insert job",
			setup: func() {
				s.mockJson.EXPECT().MarshalString(internalTask).Return(payload, nil).Once()
				mockQuery := mocksdb.NewQuery(s.T())
				s.mockDB.EXPECT().Table(s.jobsTable).Return(mockQuery).Once()
				mockQuery.EXPECT().Insert(&models.Job{
					Queue:       queue,
					Payload:     payload,
					AvailableAt: carbon.NewDateTime(carbon.Now()),
					CreatedAt:   carbon.NewDateTime(carbon.Now()),
				}).Return(nil, assert.AnError).Once()
			},
			expectedError: errors.QueueFailedToInsertJobToDatabase.Args(&models.Job{
				Queue:       queue,
				Payload:     payload,
				AvailableAt: carbon.NewDateTime(carbon.Now()),
				CreatedAt:   carbon.NewDateTime(carbon.Now()),
			}, assert.AnError),
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			test.setup()

			database := &Database{
				db:        s.mockDB,
				jobStorer: s.mockJobStorer,
				json:      s.mockJson,
				jobsTable: s.jobsTable,
			}

			err := database.Push(task, queue)

			s.Equal(test.expectedError, err)
		})
	}
}

func (s *DatabaseTestSuite) TestIsAvailable() {
	now := carbon.Now()
	carbon.SetTestNow(now)

	mockQuery := mocksdb.NewQuery(s.T())
	mockQuery.EXPECT().WhereNull("reserved_at").Return(mockQuery).Once()
	mockQuery.EXPECT().Where("available_at <= ?", now).Return(mockQuery).Once()

	database := &Database{}
	result := database.isAvailable(mockQuery)

	s.Equal(mockQuery, result)
}
