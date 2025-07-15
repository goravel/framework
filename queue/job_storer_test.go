package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
)

// MockJob is a mock implementation of the queue.Job interface for testing
type MockJob struct {
	signature string
	handleErr error
	args      []any
}

func (j *MockJob) Signature() string {
	return j.signature
}

func (j *MockJob) Handle(args ...any) error {
	j.args = args
	return j.handleErr
}

type JobRepositoryTestSuite struct {
	suite.Suite
	jobStorer *JobStorer
}

func TestJobRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(JobRepositoryTestSuite))
}

func (s *JobRepositoryTestSuite) SetupTest() {
	s.jobStorer = NewJobStorer()
}

func (s *JobRepositoryTestSuite) TestNewJobRepository() {
	repo := NewJobStorer()
	s.NotNil(repo)
}

func (s *JobRepositoryTestSuite) TestRegister() {
	// Create mock jobs
	job1 := &MockJob{signature: "job1"}
	job2 := &MockJob{signature: "job2"}

	// Register jobs
	s.jobStorer.Register([]queue.Job{job1, job2})

	// Verify jobs were registered
	allJobs := s.jobStorer.All()
	s.Len(allJobs, 2)

	// Check if both jobs are in the repository
	foundJob1 := false
	foundJob2 := false

	for _, job := range allJobs {
		if job.Signature() == "job1" {
			foundJob1 = true
		}
		if job.Signature() == "job2" {
			foundJob2 = true
		}
	}

	s.True(foundJob1)
	s.True(foundJob2)
}

func (s *JobRepositoryTestSuite) TestGet() {
	// Create and register a mock job
	job := &MockJob{signature: "test_job"}
	s.jobStorer.Register([]queue.Job{job})

	// Test getting an existing job
	retrievedJob, err := s.jobStorer.Get("test_job")
	s.NoError(err)
	s.Equal("test_job", retrievedJob.Signature())

	// Test getting a non-existent job
	_, err = s.jobStorer.Get("non_existent_job")
	s.Error(err)
	s.Equal(errors.QueueJobNotFound.Args("non_existent_job"), err)
}

func (s *JobRepositoryTestSuite) TestCall() {
	// Create and register a mock job
	job := &MockJob{signature: "test_job"}
	s.jobStorer.Register([]queue.Job{job})

	// Test calling an existing job
	args := []any{"arg1", "arg2"}
	err := s.jobStorer.Call("test_job", args)
	s.NoError(err)
	s.Equal(args, job.args)

	// Test calling a non-existent job
	err = s.jobStorer.Call("non_existent_job", args)
	s.Error(err)
	s.Equal(errors.QueueJobNotFound.Args("non_existent_job"), err)

	// Test calling a job that returns an error
	errorJob := &MockJob{signature: "error_job", handleErr: assert.AnError}
	s.jobStorer.Register([]queue.Job{errorJob})
	err = s.jobStorer.Call("error_job", args)
	s.Error(err)
	s.Equal(assert.AnError, err)
}

func (s *JobRepositoryTestSuite) TestAll() {
	// Initially, there should be no jobs
	s.Empty(s.jobStorer.All())

	// Register some jobs
	job1 := &MockJob{signature: "job1"}
	job2 := &MockJob{signature: "job2"}
	job3 := &MockJob{signature: "job3"}

	s.jobStorer.Register([]queue.Job{job1, job2, job3})

	// Get all jobs
	allJobs := s.jobStorer.All()

	// Verify the number of jobs
	s.Len(allJobs, 3)

	// Verify all jobs are present
	signatures := make(map[string]bool)
	for _, job := range allJobs {
		signatures[job.Signature()] = true
	}

	s.True(signatures["job1"])
	s.True(signatures["job2"])
	s.True(signatures["job3"])
}
