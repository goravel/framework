package queue

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	mocksdb "github.com/goravel/framework/mocks/database/db"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mockslog "github.com/goravel/framework/mocks/log"
	mocksqueue "github.com/goravel/framework/mocks/queue"
	"github.com/goravel/framework/queue/models"
	"github.com/goravel/framework/queue/utils"
	"github.com/goravel/framework/support/carbon"
)

type WorkerTestSuite struct {
	suite.Suite
	mockConfig *mocksqueue.Config
	mockDB     *mocksdb.DB
	mockDriver *mocksqueue.Driver
	mockLog    *mockslog.Log
	mockJob    *mocksqueue.JobStorer
	mockJson   *mocksfoundation.Json
	worker     *Worker
}

func TestWorkerTestSuite(t *testing.T) {
	suite.Run(t, new(WorkerTestSuite))
}

func (s *WorkerTestSuite) SetupTest() {
	s.mockConfig = mocksqueue.NewConfig(s.T())
	s.mockDB = mocksdb.NewDB(s.T())
	s.mockDriver = mocksqueue.NewDriver(s.T())
	s.mockLog = mockslog.NewLog(s.T())
	s.mockJob = mocksqueue.NewJobStorer(s.T())
	s.mockJson = mocksfoundation.NewJson(s.T())

	s.worker = &Worker{
		config: s.mockConfig,
		db:     s.mockDB,
		driver: s.mockDriver,
		job:    s.mockJob,
		json:   s.mockJson,
		log:    s.mockLog,

		connection: "sync",
		queue:      "default",
		concurrent: 1,
		debug:      true,

		currentDelay:  1 * time.Second,
		failedJobChan: make(chan models.FailedJob, 1),
		maxDelay:      32 * time.Second,
	}
}

func (s *WorkerTestSuite) TestNewWorker() {
	s.Run("happy path", func() {
		s.mockConfig.EXPECT().Driver("sync").Return(contractsqueue.DriverSync).Once()
		s.mockConfig.EXPECT().Debug().Return(true).Once()
		worker, err := NewWorker(s.mockConfig, s.mockDB, s.mockJob, s.mockJson, s.mockLog, "sync", "default", 2)

		s.NotNil(worker)
		s.NoError(err)
	})

	s.Run("failed to create driver", func() {
		s.mockConfig.EXPECT().Driver("sync").Return("unknown").Once()
		worker, err := NewWorker(s.mockConfig, s.mockDB, s.mockJob, s.mockJson, s.mockLog, "sync", "default", 2)
		s.Nil(worker)
		s.Equal(errors.QueueDriverNotSupported.Args("unknown"), err)
	})
}

func (s *WorkerTestSuite) Test_call() {

	carbon.SetTestNow(carbon.FromStdTime(time.Now()))
	defer carbon.ClearTestNow()

	task := contractsqueue.Task{
		ChainJob: contractsqueue.ChainJob{
			Job: &TestJobOne{},
			Args: []contractsqueue.Arg{
				{
					Type:  "string",
					Value: "test",
				},
			},
			Delay: carbon.Now().AddSecond().StdTime(),
		},
		UUID: "test",
		Chain: []contractsqueue.ChainJob{
			{
				Job: &TestJobTwo{},
				Args: []contractsqueue.Arg{
					{
						Type:  "int",
						Value: 1,
					},
				},
				Delay: carbon.Now().AddSecond().StdTime(),
			},
		},
	}

	s.Run("happy path", func() {
		s.SetupTest()

		s.mockJob.EXPECT().Call(task.Job.Signature(), utils.ConvertArgs(task.Args)).Return(nil).Once()

		err := s.worker.call(task)
		s.NoError(err)
	})

	s.Run("failed to call job", func() {
		s.SetupTest()

		s.mockJob.EXPECT().Call(task.Job.Signature(), utils.ConvertArgs(task.Args)).Return(assert.AnError).Once()
		s.mockJson.EXPECT().MarshalString(utils.Task{
			Job: utils.Job{
				Signature: task.Job.Signature(),
				Args:      task.Args,
				Delay:     &task.Delay,
			},
			UUID: "test",
			Chain: []utils.Job{
				{
					Signature: task.Chain[0].Job.Signature(),
					Args:      task.Chain[0].Args,
					Delay:     &task.Chain[0].Delay,
				},
			},
		}).Return("{\"signature\":\"test_job_one\",\"args\":[{\"type\":\"string\",\"value\":\"test\"}],\"delay\":null,\"uuid\":\"test\",\"chain\":[{\"signature\":\"test_job_two\",\"args\":[{\"type\":\"int\",\"value\":1}],\"delay\":null,\"uuid\":\"test\",\"chain\":[]}]}", nil).Once()

		err := s.worker.call(task)
		s.Equal(errors.QueueFailedToCallJob, err)
	})
}

func (s *WorkerTestSuite) Test_logFailedJob() {

	failedJob := models.FailedJob{
		UUID: "test",
	}

	s.Run("happy path", func() {
		s.SetupTest()

		s.mockConfig.EXPECT().FailedDatabase().Return("mysql").Once()
		s.mockConfig.EXPECT().FailedTable().Return("failed_jobs").Once()
		s.mockDB.EXPECT().Connection("mysql").Return(s.mockDB).Once()
		mockQuery := mocksdb.NewQuery(s.T())
		s.mockDB.EXPECT().Table("failed_jobs").Return(mockQuery).Once()
		mockQuery.EXPECT().Insert(&failedJob).Return(nil, nil).Once()

		s.worker.logFailedJob(failedJob)
	})

	s.Run("failed to insert failed job", func() {
		s.SetupTest()

		s.mockConfig.EXPECT().FailedDatabase().Return("mysql").Once()
		s.mockConfig.EXPECT().FailedTable().Return("failed_jobs").Once()
		s.mockDB.EXPECT().Connection("mysql").Return(s.mockDB).Once()
		mockQuery := mocksdb.NewQuery(s.T())
		s.mockDB.EXPECT().Table("failed_jobs").Return(mockQuery).Once()
		mockQuery.EXPECT().Insert(&failedJob).Return(nil, assert.AnError).Once()
		s.mockLog.EXPECT().Error(errors.QueueFailedToSaveFailedJob.Args(assert.AnError, failedJob)).Once()

		s.worker.logFailedJob(failedJob)
	})

	s.Run("db is nil", func() {
		s.SetupTest()

		s.mockConfig.EXPECT().FailedDatabase().Return("mysql").Once()
		s.mockConfig.EXPECT().FailedTable().Return("failed_jobs").Once()
		s.mockLog.EXPECT().Error(errors.QueueJobFailed.Args(failedJob)).Once()

		s.worker.db = nil
		s.worker.logFailedJob(failedJob)
	})

	s.Run("FailedDatabase is empty", func() {
		s.SetupTest()

		s.mockConfig.EXPECT().FailedDatabase().Return("").Once()
		s.mockConfig.EXPECT().FailedTable().Return("failed_jobs").Once()
		s.mockLog.EXPECT().Error(errors.QueueJobFailed.Args(failedJob)).Once()

		s.worker.logFailedJob(failedJob)
	})

	s.Run("FailedTable is empty", func() {
		s.SetupTest()

		s.mockConfig.EXPECT().FailedDatabase().Return("mysql").Once()
		s.mockConfig.EXPECT().FailedTable().Return("").Once()
		s.mockLog.EXPECT().Error(errors.QueueJobFailed.Args(failedJob)).Once()

		s.worker.logFailedJob(failedJob)
	})
}

func (s *WorkerTestSuite) Test_run() {
	carbon.SetTestNow(carbon.FromStdTime(time.Now()))
	defer carbon.ClearTestNow()

	connection := "sync"
	queue := "default"
	testJobErr := &TestJobErr{}
	testJobOne := &TestJobOne{}
	testJobTwo := &TestJobTwo{}

	errorTask := contractsqueue.Task{
		ChainJob: contractsqueue.ChainJob{
			Job: testJobErr,
		},
		UUID: "test",
	}
	errorInternalTask := utils.Task{
		Job: utils.Job{
			Signature: testJobErr.Signature(),
		},
		UUID: "test",
	}

	failedJob := &models.FailedJob{
		UUID:       errorTask.UUID,
		Connection: connection,
		Queue:      queue,
		Payload:    "{\"signature\":\"test_job_err\",\"args\":null,\"delay\":null,\"uuid\":\"test\",\"chain\":[]}",
		Exception:  assert.AnError.Error(),
		FailedAt:   carbon.NewDateTime(carbon.Now()),
	}

	s.Run("no job found", func() {
		s.SetupTest()
		s.mockDriver.EXPECT().Pop(queue).Return(nil, errors.QueueDriverNoJobFound).Once()

		go func() {
			err := s.worker.run()
			s.NoError(err)
		}()

		time.Sleep(500 * time.Millisecond)

		s.NoError(s.worker.Shutdown())
	})

	s.Run("failed to pop job", func() {
		s.SetupTest()
		s.mockDriver.EXPECT().Pop(queue).Return(nil, assert.AnError).Once()

		s.mockLog.EXPECT().Error(errors.QueueDriverFailedToPop.Args(queue, assert.AnError)).Once()

		go func() {
			err := s.worker.run()
			s.NoError(err)
		}()

		time.Sleep(500 * time.Millisecond)

		s.NoError(s.worker.Shutdown())
	})

	s.Run("job failed, insert failed job", func() {
		s.SetupTest()

		// run
		mockReservedJob := mocksqueue.NewReservedJob(s.T())
		s.mockDriver.EXPECT().Pop(queue).Return(mockReservedJob, nil).Once()
		mockReservedJob.EXPECT().Task().Return(errorTask).Once()

		// call
		s.mockJob.EXPECT().Call(errorTask.Job.Signature(), make([]any, 0)).Return(assert.AnError).Once()
		s.mockJson.EXPECT().MarshalString(errorInternalTask).Return("{\"signature\":\"test_job_err\",\"args\":null,\"delay\":null,\"uuid\":\"test\",\"chain\":[]}", nil).Once()

		// run
		mockReservedJob.EXPECT().Delete().Return(nil).Once()
		s.mockDriver.EXPECT().Pop(queue).Return(nil, errors.QueueDriverNoJobFound).Once()

		// logFailedJob
		s.mockConfig.EXPECT().FailedDatabase().Return("mysql").Once()
		s.mockConfig.EXPECT().FailedTable().Return("failed_jobs").Once()
		s.mockDB.EXPECT().Connection("mysql").Return(s.mockDB).Once()
		mockQuery := mocksdb.NewQuery(s.T())
		s.mockDB.EXPECT().Table("failed_jobs").Return(mockQuery).Once()
		mockQuery.EXPECT().Insert(failedJob).Return(nil, nil).Once()

		go func() {
			err := s.worker.run()
			s.NoError(err)
		}()

		time.Sleep(500 * time.Millisecond)

		s.NoError(s.worker.Shutdown())
	})

	s.Run("failed to insert failed job", func() {
		s.SetupTest()

		// run
		mockReservedJob := mocksqueue.NewReservedJob(s.T())
		s.mockDriver.EXPECT().Pop(queue).Return(mockReservedJob, nil).Once()
		mockReservedJob.EXPECT().Task().Return(errorTask).Once()

		// call
		s.mockJob.EXPECT().Call(errorTask.Job.Signature(), make([]any, 0)).Return(assert.AnError).Once()
		s.mockJson.EXPECT().MarshalString(errorInternalTask).Return("{\"signature\":\"test_job_err\",\"args\":null,\"delay\":null,\"uuid\":\"test\",\"chain\":[]}", nil).Once()

		// run
		mockReservedJob.EXPECT().Delete().Return(nil).Once()
		s.mockDriver.EXPECT().Pop(queue).Return(nil, errors.QueueDriverNoJobFound).Once()

		// logFailedJob
		s.mockConfig.EXPECT().FailedDatabase().Return("mysql").Once()
		s.mockConfig.EXPECT().FailedTable().Return("failed_jobs").Once()
		s.mockDB.EXPECT().Connection("mysql").Return(s.mockDB).Once()
		mockQuery := mocksdb.NewQuery(s.T())
		s.mockDB.EXPECT().Table("failed_jobs").Return(mockQuery).Once()
		mockQuery.EXPECT().Insert(failedJob).Return(nil, assert.AnError).Once()
		s.mockLog.EXPECT().Error(errors.QueueFailedToSaveFailedJob.Args(assert.AnError, failedJob)).Once()

		go func() {
			err := s.worker.run()
			s.NoError(err)
		}()

		time.Sleep(500 * time.Millisecond)

		s.NoError(s.worker.Shutdown())
	})

	s.Run("chain job failed, insert failed job", func() {
		s.SetupTest()

		args := []contractsqueue.Arg{
			{
				Type:  "string",
				Value: "test",
			},
		}
		errorTaskWithChain := contractsqueue.Task{
			ChainJob: contractsqueue.ChainJob{
				Job: testJobOne,
			},
			UUID: "test",
			Chain: []contractsqueue.ChainJob{
				{
					Job: testJobErr,
					Args: []contractsqueue.Arg{
						{
							Type:  "string",
							Value: "test",
						},
					},
				},
			},
		}

		// run
		mockReservedJob := mocksqueue.NewReservedJob(s.T())
		s.mockDriver.EXPECT().Pop(queue).Return(mockReservedJob, nil).Once()
		mockReservedJob.EXPECT().Task().Return(errorTaskWithChain).Once()

		// call
		s.mockJob.EXPECT().Call(errorTaskWithChain.Job.Signature(), make([]any, 0)).Return(nil).Once()
		s.mockJob.EXPECT().Call(errorTaskWithChain.Chain[0].Job.Signature(), utils.ConvertArgs(args)).Return(assert.AnError).Once()
		s.mockJson.EXPECT().MarshalString(utils.Task{
			Job: utils.Job{
				Signature: errorTaskWithChain.Chain[0].Job.Signature(),
				Args:      args,
			},
			UUID: "test",
		}).Return("{\"signature\":\"test_job_err\",\"args\":[{\"type\":\"string\",\"value\":\"test\"}],\"delay\":null,\"uuid\":\"test\",\"chain\":[]}", nil).Once()

		// run
		mockReservedJob.EXPECT().Delete().Return(nil).Once()
		s.mockDriver.EXPECT().Pop(queue).Return(nil, errors.QueueDriverNoJobFound).Once()

		// logFailedJob
		s.mockConfig.EXPECT().FailedDatabase().Return("mysql").Once()
		s.mockConfig.EXPECT().FailedTable().Return("failed_jobs").Once()
		s.mockDB.EXPECT().Connection("mysql").Return(s.mockDB).Once()
		mockQuery := mocksdb.NewQuery(s.T())
		s.mockDB.EXPECT().Table("failed_jobs").Return(mockQuery).Once()
		mockQuery.EXPECT().Insert(&models.FailedJob{
			UUID:       errorTaskWithChain.UUID,
			Connection: connection,
			Queue:      queue,
			Payload:    "{\"signature\":\"test_job_err\",\"args\":[{\"type\":\"string\",\"value\":\"test\"}],\"delay\":null,\"uuid\":\"test\",\"chain\":[]}",
			Exception:  assert.AnError.Error(),
			FailedAt:   carbon.NewDateTime(carbon.Now()),
		}).Return(nil, nil).Once()

		go func() {
			err := s.worker.run()
			s.NoError(err)
		}()

		time.Sleep(500 * time.Millisecond)

		s.NoError(s.worker.Shutdown())
	})

	s.Run("happy path", func() {
		s.SetupTest()

		successTask := contractsqueue.Task{
			ChainJob: contractsqueue.ChainJob{
				Job:   testJobOne,
				Args:  testArgs,
				Delay: carbon.Now().AddSecond().StdTime(),
			},
			UUID:  "test",
			Chain: []contractsqueue.ChainJob{},
		}

		// run
		mockReservedJob := mocksqueue.NewReservedJob(s.T())
		s.mockDriver.EXPECT().Pop(queue).Return(mockReservedJob, nil).Once()
		mockReservedJob.EXPECT().Task().Return(successTask).Once()

		// call
		s.mockJob.EXPECT().Call(successTask.Job.Signature(), utils.ConvertArgs(testArgs)).Return(nil).Once()

		// run
		mockReservedJob.EXPECT().Delete().Return(nil).Once()
		s.mockDriver.EXPECT().Pop(queue).Return(nil, errors.QueueDriverNoJobFound).Once()

		go func() {
			err := s.worker.run()
			s.NoError(err)
		}()

		time.Sleep(1500 * time.Millisecond)

		s.NoError(s.worker.Shutdown())
	})

	s.Run("happy path with chain", func() {
		s.SetupTest()

		successTaskWithChain := contractsqueue.Task{
			ChainJob: contractsqueue.ChainJob{
				Job: testJobOne,
			},
			UUID: "test",
			Chain: []contractsqueue.ChainJob{
				{
					Job:  testJobTwo,
					Args: testArgs,
				},
			},
		}

		// run
		mockReservedJob := mocksqueue.NewReservedJob(s.T())
		s.mockDriver.EXPECT().Pop(queue).Return(mockReservedJob, nil).Once()
		mockReservedJob.EXPECT().Task().Return(successTaskWithChain).Once()

		// call
		s.mockJob.EXPECT().Call(successTaskWithChain.Job.Signature(), make([]any, 0)).Return(nil).Once()
		s.mockJob.EXPECT().Call(successTaskWithChain.Chain[0].Job.Signature(), utils.ConvertArgs(testArgs)).Return(nil).Once()

		// run
		mockReservedJob.EXPECT().Delete().Return(nil).Once()
		s.mockDriver.EXPECT().Pop(queue).Return(nil, errors.QueueDriverNoJobFound).Once()

		go func() {
			err := s.worker.run()
			s.NoError(err)
		}()

		time.Sleep(500 * time.Millisecond)

		s.NoError(s.worker.Shutdown())
	})
}

func (s *WorkerTestSuite) TestRunWithSyncDriver() {
	s.mockDriver.EXPECT().Driver().Return(contractsqueue.DriverSync).Once()

	err := s.worker.Run()
	s.NoError(err)
}

func (s *WorkerTestSuite) TestShutdown() {
	s.worker.isShutdown.Store(false)

	err := s.worker.Shutdown()
	s.NoError(err)
	s.True(s.worker.isShutdown.Load())
}
