package queue

import (
	"fmt"
	"testing"
	"time"

	machinerylog "github.com/RichardKnop/machinery/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mockslog "github.com/goravel/framework/mocks/log"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/docker"
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/testing/utils"
)

type MachineryTestSuite struct {
	suite.Suite
	docker    *docker.Redis
	machinery *Machinery
}

func TestMachineryTestSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("skip on windows")
	}

	suite.Run(t, new(MachineryTestSuite))
}

func (s *MachineryTestSuite) SetupSuite() {
	redis := docker.NewRedis()
	if err := redis.Build(); err != nil {
		s.T().Fatalf("failed to build redis docker: %v", err)
	}

	s.docker = redis
}

func (s *MachineryTestSuite) TearDownSuite() {
	s.NoError(s.docker.Shutdown())
}

func (s *MachineryTestSuite) SetupTest() {
	testJobOne = nil
	testJobTwo = nil
	debug := true
	log := utils.NewTestLog()

	machinerylog.DEBUG = NewDebug(debug, log)
	machinerylog.INFO = NewInfo(debug, log)
	machinerylog.WARNING = NewWarning(debug, log)
	machinerylog.ERROR = NewError(debug, log)
	machinerylog.FATAL = NewFatal(debug, log)

	s.machinery = &Machinery{
		appName:       "goravel",
		log:           log,
		redisDatabase: 0,
		redisDSN:      fmt.Sprintf("localhost:%d", s.docker.Config().Port),
	}
}

func (s *MachineryTestSuite) Test_Driver() {
	s.Equal(contractsqueue.DriverMachinery, s.machinery.Driver())
}

func (s *MachineryTestSuite) Test_NewMachinery() {
	mockLog := mockslog.NewLog(s.T())
	mockConfig := mocksconfig.NewConfig(s.T())
	mockConfig.EXPECT().GetString("app.name").Return("goravel").Once()
	mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
	mockConfig.EXPECT().GetString("queue.connections.machinery.connection").Return("machinery").Once()
	mockConfig.EXPECT().GetString("database.redis.machinery.host").Return("localhost").Once()
	mockConfig.EXPECT().GetString("database.redis.machinery.password").Return("").Once()
	mockConfig.EXPECT().GetInt("database.redis.machinery.port").Return(s.docker.Config().Port).Once()
	mockConfig.EXPECT().GetInt("database.redis.machinery.database").Return(0).Once()

	machinery := NewMachinery(mockConfig, mockLog, "machinery")

	s.NotNil(machinery)
	s.Equal("goravel", machinery.appName)
	s.Equal(mockLog, machinery.log)
	s.Equal(0, machinery.redisDatabase)
	s.Equal(fmt.Sprintf("localhost:%d", s.docker.Config().Port), machinery.redisDSN)
}

func (s *MachineryTestSuite) Test_PushAndRun() {
	queue := "default"
	err := s.machinery.Push(contractsqueue.Task{
		ChainJob: contractsqueue.ChainJob{
			Job: &TestJobOne{},
			Args: []contractsqueue.Arg{
				{Type: "string", Value: "a"},
				{Type: "int", Value: 1},
				{Type: "[]string", Value: []string{"b", "c"}},
				{Type: "[]int", Value: []int{1, 2, 3}},
			},
		},
	}, queue)

	s.Require().NoError(err)

	jobs := []contractsqueue.Job{&TestJobOne{}, &TestJobTwo{}, &TestJobErr{}}
	worker, err := s.machinery.Run(jobs, queue, 1)
	s.Require().NoError(err)

	defer worker.Quit()

	time.Sleep(time.Second)

	s.Equal("a", testJobOne[0])
	s.Equal(1, testJobOne[1])
	s.Equal([]string{"b", "c"}, testJobOne[2])
	s.Equal([]int{1, 2, 3}, testJobOne[3])
}

func (s *MachineryTestSuite) Test_PushAndRunWithDelay() {
	queue := "default"
	err := s.machinery.Push(contractsqueue.Task{
		ChainJob: contractsqueue.ChainJob{
			Job: &TestJobOne{},
			Args: []contractsqueue.Arg{
				{Type: "string", Value: "a"},
				{Type: "int", Value: 1},
				{Type: "[]string", Value: []string{"b", "c"}},
				{Type: "[]int", Value: []int{1, 2, 3}},
			},
			Delay: carbon.Now().AddSecond().StdTime(),
		},
	}, queue)

	s.Require().NoError(err)

	jobs := []contractsqueue.Job{&TestJobOne{}, &TestJobTwo{}, &TestJobErr{}}
	worker, err := s.machinery.Run(jobs, queue, 1)
	s.Require().NoError(err)

	defer worker.Quit()

	time.Sleep(time.Second)

	s.Len(testJobOne, 0)

	time.Sleep(time.Second)

	s.Equal("a", testJobOne[0])
	s.Equal(1, testJobOne[1])
	s.Equal([]string{"b", "c"}, testJobOne[2])
	s.Equal([]int{1, 2, 3}, testJobOne[3])
}

func (s *MachineryTestSuite) Test_PushChainAndRun() {
	queue := "default"
	err := s.machinery.Push(contractsqueue.Task{
		ChainJob: contractsqueue.ChainJob{
			Job: &TestJobOne{},
			Args: []contractsqueue.Arg{
				{Type: "string", Value: "a"},
				{Type: "int", Value: 1},
				{Type: "[]string", Value: []string{"b", "c"}},
				{Type: "[]int", Value: []int{1, 2, 3}},
			},
		},
		Chain: []contractsqueue.ChainJob{
			{
				Job: &TestJobTwo{},
				Args: []contractsqueue.Arg{
					{Type: "string", Value: "b"},
				},
				Delay: carbon.Now().AddSecond().StdTime(),
			},
		},
	}, queue)

	s.Require().NoError(err)

	jobs := []contractsqueue.Job{&TestJobOne{}, &TestJobTwo{}, &TestJobErr{}}
	worker, err := s.machinery.Run(jobs, queue, 1)
	s.Require().NoError(err)

	defer worker.Quit()

	time.Sleep(time.Second)

	s.Equal("a", testJobOne[0])
	s.Equal(1, testJobOne[1])
	s.Equal([]string{"b", "c"}, testJobOne[2])
	s.Equal([]int{1, 2, 3}, testJobOne[3])
	s.Len(testJobTwo, 0)

	time.Sleep(time.Second)

	s.Equal("b", testJobTwo[0])
}

func (s *MachineryTestSuite) Test_queueKey() {
	s.Equal("goravel_queues:default", s.machinery.queueKey("default"))
}

func Test_jobs2Tasks(t *testing.T) {
	// Test successful conversion
	jobs := []contractsqueue.Job{&TestJobOne{}, &TestJobTwo{}}
	tasks, err := jobs2Tasks(jobs)
	assert.NoError(t, err)
	assert.Len(t, tasks, 2)
	assert.NotNil(t, tasks[(&TestJobOne{}).Signature()])
	assert.NotNil(t, tasks[(&TestJobTwo{}).Signature()])

	// Test empty signature error
	emptyJob := &TestJobEmptySignature{}
	jobs = []contractsqueue.Job{emptyJob}
	tasks, err = jobs2Tasks(jobs)
	assert.Error(t, err)
	assert.Equal(t, errors.QueueEmptyJobSignature, err)
	assert.Nil(t, tasks)

	// Test duplicate signature error
	duplicateJobs := []contractsqueue.Job{&TestJobOne{}, &TestJobOne{}}
	tasks, err = jobs2Tasks(duplicateJobs)
	assert.Error(t, err)
	assert.Equal(t, errors.QueueDuplicateJobSignature.Args((&TestJobOne{}).Signature()), err)
	assert.Nil(t, tasks)
}

// TestJobEmptySignature is a test job with empty signature
type TestJobEmptySignature struct{}

func (j *TestJobEmptySignature) Signature() string {
	return ""
}

func (j *TestJobEmptySignature) Handle(args ...any) error {
	return nil
}
