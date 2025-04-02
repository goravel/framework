package queue

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/queue"
)

type JobTestSuite struct {
	suite.Suite
	jobManager *JobRespository
}

func TestJobTestSuite(t *testing.T) {
	suite.Run(t, new(JobTestSuite))
}

func (s *JobTestSuite) SetupTest() {
	s.jobManager = NewJobRespository()
}

func (s *JobTestSuite) RegisterJobsSuccessfully() {
	s.jobManager.Register([]queue.Job{
		&MockJob{signature: "job1"},
		&MockJob{signature: "job2"},
	})

	registeredJobs := s.jobManager.All()
	s.Len(registeredJobs, 2)
}

func (s *JobTestSuite) CallRegisteredJobSuccessfully() {
	job := &MockJob{signature: "job1"}
	s.jobManager.Register([]queue.Job{job})

	err := s.jobManager.Call("job1", []any{"arg1"})
	s.NoError(err)
	s.True(job.called)
}

func (s *JobTestSuite) CallUnregisteredJobFails() {
	err := s.jobManager.Call("nonexistent", []any{"arg1"})
	s.Error(err)
}

func (s *JobTestSuite) GetRegisteredJobSuccessfully() {
	job := &MockJob{signature: "job1"}
	s.jobManager.Register([]queue.Job{job})

	retrievedJob, err := s.jobManager.Get("job1")
	s.NoError(err)
	s.Equal(job, retrievedJob)
}

func (s *JobTestSuite) GetUnregisteredJobFails() {
	_, err := s.jobManager.Get("nonexistent")
	s.Error(err)
}

type MockJob struct {
	signature string
	called    bool
}

func (m *MockJob) Signature() string {
	return m.signature
}

func (m *MockJob) Handle(args ...any) error {
	m.called = true
	return nil
}
