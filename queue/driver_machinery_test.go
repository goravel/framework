package queue

import (
	"testing"
	"time"

	"github.com/RichardKnop/machinery/v2/tasks"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/queue"
	mocksconfig "github.com/goravel/framework/mocks/config"
	"github.com/goravel/framework/support/convert"
	"github.com/goravel/framework/support/docker"
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/testing/utils"
)

type MachineryTestSuite struct {
	suite.Suite
	dockerPort int
	machinery  *Machinery
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

	s.dockerPort = redis.Config().Port
}

func (s *MachineryTestSuite) SetupTest() {
	testJobOne = nil

	mockConfig := mocksconfig.NewConfig(s.T())
	mockConfig.EXPECT().GetString("app.name").Return("goravel").Once()
	mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
	mockConfig.EXPECT().GetString("queue.connections.redis.connection").Return("redis").Once()
	mockConfig.EXPECT().GetString("database.redis.redis.host").Return("localhost").Once()
	mockConfig.EXPECT().GetString("database.redis.redis.password").Return("").Once()
	mockConfig.EXPECT().GetInt("database.redis.redis.port").Return(s.dockerPort).Once()
	mockConfig.EXPECT().GetInt("database.redis.redis.database").Return(0).Once()

	jobs := []queue.Job{&TestJobOne{}, &TestJobTwo{}, &TestJobErr{}}

	s.machinery = NewMachinery(mockConfig, utils.NewTestLog(), jobs, "redis", "default", 1)
}

func (s *MachineryTestSuite) TestDispatch() {
	job := &TestJobOne{}
	s.machinery.Server().SendTask(&tasks.Signature{
		Name: job.Signature(),
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: "a",
			},
			{
				Type:  "int",
				Value: 1,
			},
			{
				Type:  "[]string",
				Value: []string{"b", "c"},
			},
			{
				Type:  "[]int",
				Value: []int{1, 2, 3},
			},
		},
	})

	s.Require().True(s.machinery.ExistTasks())

	worker, err := s.machinery.Run()
	s.Require().NoError(err)

	defer worker.Quit()

	time.Sleep(time.Second)

	s.Equal("a", testJobOne[0])
	s.Equal(1, testJobOne[1])
	s.Equal([]string{"b", "c"}, testJobOne[2])
	s.Equal([]int{1, 2, 3}, testJobOne[3])
}

func (s *MachineryTestSuite) TestDelay() {
	job := &TestJobOne{}
	s.machinery.Server().SendTask(&tasks.Signature{
		Name: job.Signature(),
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: "a",
			},
			{
				Type:  "int",
				Value: 1,
			},
			{
				Type:  "[]string",
				Value: []string{"b", "c"},
			},
			{
				Type:  "[]int",
				Value: []int{1, 2, 3},
			},
		},
		ETA: convert.Pointer(time.Now().Add(time.Second * 1)),
	})

	s.Require().True(s.machinery.ExistTasks())

	worker, err := s.machinery.Run()
	s.Require().NoError(err)

	defer worker.Quit()

	time.Sleep(2 * time.Second)

	s.Equal("a", testJobOne[0])
	s.Equal(1, testJobOne[1])
	s.Equal([]string{"b", "c"}, testJobOne[2])
	s.Equal([]int{1, 2, 3}, testJobOne[3])
}
