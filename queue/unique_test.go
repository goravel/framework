package queue

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/goravel/framework/cache"
	contractscache "github.com/goravel/framework/contracts/cache"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksqueue "github.com/goravel/framework/mocks/queue"
	"github.com/goravel/framework/queue/utils"
)

type uniqueTestJob struct {
	uniqueFor time.Duration
}

func (r *uniqueTestJob) Signature() string {
	return "unique_test_job"
}

func (r *uniqueTestJob) Handle(...any) error {
	return nil
}

func (r *uniqueTestJob) UniqueID(args ...any) string {
	if len(args) == 0 {
		return ""
	}

	return fmt.Sprint(args[0])
}

func (r *uniqueTestJob) UniqueFor() time.Duration {
	return r.uniqueFor
}

// uniqueTestCache is a cache with a single memory store.
type uniqueTestCache struct {
	*cache.Memory
}

func (r uniqueTestCache) Store(string) contractscache.Driver {
	return r.Memory
}

func newUniqueTestCache(t *testing.T) contractscache.Cache {
	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("cache.prefix").Return("").Once()

	memory, err := cache.NewMemory(mockConfig)
	require.NoError(t, err)

	return uniqueTestCache{Memory: memory}
}

func newUniquePendingJob(cache contractscache.Cache, driverCreator contractsqueue.DriverCreator, job contractsqueue.Job, id string) *PendingJob {
	var args []contractsqueue.Arg
	if id != "" {
		args = []contractsqueue.Arg{{Type: "string", Value: id}}
	}

	return &PendingJob{
		cache:         cache,
		connection:    "redis",
		driverCreator: driverCreator,
		queue:         "default",
		task: contractsqueue.Task{
			UUID:     id,
			ChainJob: contractsqueue.ChainJob{Job: job, Args: args},
		},
	}
}

func TestDispatchUniqueJob(t *testing.T) {
	memory := newUniqueTestCache(t)
	mockDriverCreator := mocksqueue.NewDriverCreator(t)
	mockDriver := mocksqueue.NewDriver(t)
	mockDriverCreator.EXPECT().Create("redis").Return(mockDriver, nil)
	mockDriver.EXPECT().Driver().Return("redis")

	job := &uniqueTestJob{}
	first := newUniquePendingJob(memory, mockDriverCreator, job, "1")
	mockDriver.EXPECT().Push(first.task, "default").Return(nil).Twice()
	other := newUniquePendingJob(memory, mockDriverCreator, job, "2")
	mockDriver.EXPECT().Push(other.task, "default").Return(nil).Once()

	assert.NoError(t, first.Dispatch())
	// A job with the same unique ID is pending, so this one is skipped.
	assert.NoError(t, newUniquePendingJob(memory, mockDriverCreator, job, "1").Dispatch())
	// Another unique ID is pushed.
	assert.NoError(t, other.Dispatch())

	// Once the first job is processed, it can be pushed again.
	releaseUniqueLock(memory, first.task)
	assert.NoError(t, first.Dispatch())
}

func TestDispatchUniqueJobWithoutID(t *testing.T) {
	memory := newUniqueTestCache(t)
	mockDriverCreator := mocksqueue.NewDriverCreator(t)
	mockDriver := mocksqueue.NewDriver(t)
	mockDriverCreator.EXPECT().Create("redis").Return(mockDriver, nil)
	mockDriver.EXPECT().Driver().Return("redis")

	first := newUniquePendingJob(memory, mockDriverCreator, &uniqueTestJob{}, "")
	mockDriver.EXPECT().Push(first.task, "default").Return(nil).Once()

	assert.NoError(t, first.Dispatch())
	// Without an ID the job is unique by its signature alone.
	assert.NoError(t, newUniquePendingJob(memory, mockDriverCreator, &uniqueTestJob{}, "").Dispatch())
}

func TestDispatchUniqueJobReleasesTheLock(t *testing.T) {
	t.Run("when the push fails", func(t *testing.T) {
		memory := newUniqueTestCache(t)
		mockDriverCreator := mocksqueue.NewDriverCreator(t)
		mockDriver := mocksqueue.NewDriver(t)
		mockDriverCreator.EXPECT().Create("redis").Return(mockDriver, nil)

		pendingJob := newUniquePendingJob(memory, mockDriverCreator, &uniqueTestJob{}, "1")
		mockDriver.EXPECT().Push(pendingJob.task, "default").Return(assert.AnError).Once()
		mockDriver.EXPECT().Push(pendingJob.task, "default").Return(nil).Once()
		mockDriver.EXPECT().Driver().Return("redis").Once()

		assert.ErrorIs(t, pendingJob.Dispatch(), assert.AnError)
		assert.NoError(t, pendingJob.Dispatch())
	})

	t.Run("when the sync driver has run the job", func(t *testing.T) {
		memory := newUniqueTestCache(t)
		mockDriverCreator := mocksqueue.NewDriverCreator(t)
		mockDriver := mocksqueue.NewDriver(t)
		mockDriverCreator.EXPECT().Create("redis").Return(mockDriver, nil)
		mockDriver.EXPECT().Driver().Return(contractsqueue.DriverSync)

		pendingJob := newUniquePendingJob(memory, mockDriverCreator, &uniqueTestJob{}, "1")
		mockDriver.EXPECT().Push(pendingJob.task, "default").Return(nil).Twice()

		assert.NoError(t, pendingJob.Dispatch())
		assert.NoError(t, pendingJob.Dispatch())
	})

	t.Run("after UniqueFor", func(t *testing.T) {
		memory := newUniqueTestCache(t)
		mockDriverCreator := mocksqueue.NewDriverCreator(t)
		mockDriver := mocksqueue.NewDriver(t)
		mockDriverCreator.EXPECT().Create("redis").Return(mockDriver, nil)
		mockDriver.EXPECT().Driver().Return("redis")

		pendingJob := newUniquePendingJob(memory, mockDriverCreator, &uniqueTestJob{uniqueFor: time.Second}, "1")
		mockDriver.EXPECT().Push(pendingJob.task, "default").Return(nil).Twice()

		assert.NoError(t, pendingJob.Dispatch())
		assert.NoError(t, pendingJob.Dispatch())

		time.Sleep(1100 * time.Millisecond)
		assert.NoError(t, pendingJob.Dispatch())
	})
}

func TestDispatchUniqueJobWithoutCache(t *testing.T) {
	mockDriverCreator := mocksqueue.NewDriverCreator(t)
	mockDriverCreator.EXPECT().Create("redis").Return(mocksqueue.NewDriver(t), nil).Once()

	err := newUniquePendingJob(nil, mockDriverCreator, &uniqueTestJob{}, "1").Dispatch()

	assert.EqualError(t, err, errors.QueueUniqueJobWithoutCache.Args("unique_test_job").Error())
}

func TestDispatchSyncIgnoresUniqueness(t *testing.T) {
	memory := newUniqueTestCache(t)
	pendingJob := newUniquePendingJob(memory, nil, &uniqueTestJob{}, "1")

	assert.True(t, memory.Lock(uniqueLockPrefix+"unique_test_job:1").Get())
	assert.NoError(t, pendingJob.DispatchSync())
}

func (s *WorkerTestSuite) Test_ProcessReleasesTheUniqueLock() {
	task := contractsqueue.Task{
		ChainJob: contractsqueue.ChainJob{
			Job:  &uniqueTestJob{},
			Args: []contractsqueue.Arg{{Type: "string", Value: "1"}},
		},
		UUID: "test",
	}
	key := uniqueLockPrefix + "unique_test_job:1"

	s.Run("after the job succeeds", func() {
		s.SetupTest()
		memory := newUniqueTestCache(s.T())
		s.worker.cache = memory
		s.True(memory.Lock(key).Get())

		mockReservedJob := mocksqueue.NewReservedJob(s.T())
		mockReservedJob.EXPECT().Task().Return(task).Once()
		s.mockJob.EXPECT().Call(task.Job.Signature(), utils.ConvertArgs(task.Args)).Return(nil).Once()
		mockReservedJob.EXPECT().Delete().Return(nil).Once()

		released, err := s.worker.Process(context.Background(), mockReservedJob)
		s.False(released)
		s.NoError(err)
		s.True(memory.Lock(key).Get())
	})

	s.Run("after the job fails for good", func() {
		s.SetupTest()
		memory := newUniqueTestCache(s.T())
		s.worker.cache = memory
		s.True(memory.Lock(key).Get())

		mockReservedJob := mocksqueue.NewReservedJob(s.T())
		mockReservedJob.EXPECT().Task().Return(task).Once()
		s.mockJob.EXPECT().Call(task.Job.Signature(), utils.ConvertArgs(task.Args)).Return(assert.AnError).Once()
		mockReservedJob.EXPECT().Attempts().Return(1).Maybe()
		mockReservedJob.EXPECT().Delete().Return(nil).Once()
		s.mockJson.EXPECT().MarshalString(mock.Anything).Return("{}", nil).Once()

		released, err := s.worker.processReservedJob(context.Background(), mockReservedJob, s.worker.queue, s.worker.sendFailedJob)
		s.False(released)
		s.ErrorIs(err, errors.QueueFailedToCallJob)
		s.True(memory.Lock(key).Get())
	})

	s.Run("not while the job waits for a retry", func() {
		s.SetupTest()
		s.worker.tries = 3
		memory := newUniqueTestCache(s.T())
		s.worker.cache = memory
		s.True(memory.Lock(key).Get())

		mockReservedJob := mocksqueue.NewReservedJob(s.T())
		mockReservedJob.EXPECT().Task().Return(task).Once()
		s.mockJob.EXPECT().Call(task.Job.Signature(), utils.ConvertArgs(task.Args)).Return(assert.AnError).Once()
		mockReservedJob.EXPECT().Attempts().Return(1).Once()
		mockReservedJob.EXPECT().Release(time.Duration(0)).Return(nil).Once()

		released, err := s.worker.Process(context.Background(), mockReservedJob)
		s.True(released)
		s.NoError(err)
		s.False(memory.Lock(key).Get())
	})
}
