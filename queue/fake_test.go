package queue

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	mocksqueue "github.com/goravel/framework/mocks/queue"
)

// newFakeTestApp returns an Application whose jobs default to the redis
// connection and the default queue. The connection is never resolved while
// the jobs are faked.
func newFakeTestApp(t *testing.T) *Application {
	mockConfig := mocksqueue.NewConfig(t)
	mockConfig.EXPECT().DefaultConnection().Return("redis").Maybe()
	mockConfig.EXPECT().DefaultQueue().Return("default").Maybe()

	return NewApplication(mockConfig, nil, nil, NewJobStorer(), nil, nil)
}

func TestFake_RecordsInsteadOfPushing(t *testing.T) {
	app := newFakeTestApp(t)
	fake := app.Fake()

	delay := time.Now().Add(time.Minute)
	args := []contractsqueue.Arg{{Type: "string", Value: "goravel"}}
	job := &TestJobOne{}
	assert.NoError(t, app.Job(job, args).OnQueue("emails").Delay(delay).Dispatch())

	assert.True(t, fake.AssertPushed(&TestJobOne{}))
	assert.True(t, fake.AssertPushed(&TestJobOne{}, func(pushed contractsqueue.Pushed) bool {
		return pushed.Connection() == "redis" && pushed.Delay().Equal(delay) && pushed.Args()[0].Value == "goravel"
	}))
	assert.False(t, fake.AssertPushed(&TestJobOne{}, func(pushed contractsqueue.Pushed) bool {
		return pushed.Connection() == "sync"
	}))
	assert.True(t, fake.AssertPushedOn("emails", &TestJobOne{}))
	assert.False(t, fake.AssertPushedOn("default", &TestJobOne{}))
	assert.True(t, fake.AssertPushedTimes(&TestJobOne{}, 1))
	assert.False(t, fake.AssertPushedTimes(&TestJobOne{}, 0))
	assert.False(t, fake.AssertPushedTimes(&TestJobOne{}, 2))
	assert.True(t, fake.AssertNotPushed(&TestJobTwo{}))
	assert.False(t, fake.AssertNotPushed(&TestJobOne{}))
	assert.False(t, fake.AssertNothingPushed())
	pushed := fake.Pushed(&TestJobOne{})
	require.Len(t, pushed, 1)
	assert.Same(t, job, pushed[0].Job())
	assert.Nil(t, fake.Pushed(nil))
	assert.Nil(t, fake.Pushed((*TestJobOne)(nil)))
	assert.Nil(t, fake.Pushed(""))
	assert.Nil(t, fake.Pushed(123))
}

func TestFake_LooksJobsUpBySignatureAndWildcard(t *testing.T) {
	app := newFakeTestApp(t)
	fake := app.Fake()

	assert.NoError(t, app.Job(&TestJobOne{}).Dispatch())
	assert.NoError(t, app.Job(&TestJobTwo{}).Dispatch())

	assert.True(t, fake.AssertPushed("test_job_one"))
	assert.True(t, fake.AssertPushedTimes("test_job_*", 2))
	assert.True(t, fake.AssertPushedTimes("*", 2))
	assert.True(t, fake.AssertNotPushed("test_job_three"))
	assert.True(t, fake.AssertNotPushed("other_*"))
}

func TestFake_CombinesTheDelays(t *testing.T) {
	app := newFakeTestApp(t)
	fake := app.Fake()

	delay := time.Now().Add(time.Minute)
	assert.NoError(t, app.Chain([]contractsqueue.ChainJob{
		{Job: &TestJobOne{}, Delay: delay},
	}).Delay(delay).Dispatch())

	assert.True(t, fake.AssertPushed(&TestJobOne{}, func(pushed contractsqueue.Pushed) bool {
		return pushed.Delay().After(delay)
	}))
}

func TestFake_Chain(t *testing.T) {
	app := newFakeTestApp(t)
	fake := app.Fake()

	assert.NoError(t, app.Chain([]contractsqueue.ChainJob{
		{Job: &TestJobOne{}},
		{Job: &TestJobTwo{}, Args: []contractsqueue.Arg{{Type: "int", Value: 1}}},
	}).Dispatch())

	// The chain is recorded as its first job, like it is pushed.
	assert.True(t, fake.AssertPushedTimes(&TestJobOne{}, 1))
	assert.True(t, fake.AssertNotPushed(&TestJobTwo{}))
	assert.True(t, fake.AssertPushed(&TestJobOne{}, func(pushed contractsqueue.Pushed) bool {
		return pushed.HasChain(&TestJobTwo{})
	}))
	assert.True(t, fake.AssertPushed(&TestJobOne{}, func(pushed contractsqueue.Pushed) bool {
		return pushed.HasChain("test_job_two")
	}))
	assert.False(t, fake.AssertPushed(&TestJobOne{}, func(pushed contractsqueue.Pushed) bool {
		return pushed.HasChain(&TestJobOne{})
	}))
	assert.False(t, fake.AssertPushed(&TestJobOne{}, func(pushed contractsqueue.Pushed) bool {
		return pushed.HasChain()
	}))
}

func TestFake_OnlyTheGivenJobs(t *testing.T) {
	for name, job := range map[string]any{
		"instance":  &TestJobOne{},
		"signature": "test_job_one",
		"wildcard":  "*_one",
	} {
		t.Run(name, func(t *testing.T) {
			testFakeOnlyTheGivenJob(t, job)
		})
	}
}

func testFakeOnlyTheGivenJob(t *testing.T, job any) {
	app := newFakeTestApp(t)
	fake := app.Fake(job)

	mockDriverCreator := mocksqueue.NewDriverCreator(t)
	mockDriver := mocksqueue.NewDriver(t)

	faked, ok := app.Job(&TestJobOne{}).(*PendingJob)
	require.True(t, ok)
	faked.driverCreator = mockDriverCreator
	pushed, ok := app.Job(&TestJobTwo{}).(*PendingJob)
	require.True(t, ok)
	pushed.driverCreator = mockDriverCreator

	mockDriverCreator.EXPECT().Create("redis").Return(mockDriver, nil).Once()
	mockDriver.EXPECT().Push(pushed.task, "default").Return(nil).Once()

	assert.NoError(t, faked.Dispatch())
	assert.NoError(t, pushed.Dispatch())

	assert.True(t, fake.AssertPushed(&TestJobOne{}))
	// Jobs that are not faked are pushed and not recorded.
	assert.True(t, fake.AssertNotPushed(&TestJobTwo{}))
}

func TestFake_DispatchSyncIsNotFaked(t *testing.T) {
	app := newFakeTestApp(t)
	fake := app.Fake()
	testJobOne = nil
	t.Cleanup(func() { testJobOne = nil })

	assert.NoError(t, app.Job(&TestJobOne{}, []contractsqueue.Arg{{Type: "string", Value: "sync"}}).DispatchSync())

	assert.Equal(t, []any{"sync"}, testJobOne)
	assert.True(t, fake.AssertNothingPushed())
}

func TestFake_InvalidJobFakesNothing(t *testing.T) {
	mockDriverCreator := mocksqueue.NewDriverCreator(t)
	mockDriver := mocksqueue.NewDriver(t)
	fake := NewFake()
	fake.enable([]any{123, ""})

	pendingJob := &PendingJob{
		connection:    "redis",
		driverCreator: mockDriverCreator,
		fake:          fake,
		queue:         "default",
		task:          contractsqueue.Task{ChainJob: contractsqueue.ChainJob{Job: &TestJobOne{}}},
	}
	mockDriverCreator.EXPECT().Create("redis").Return(mockDriver, nil).Once()
	mockDriver.EXPECT().Push(pendingJob.task, "default").Return(nil).Once()

	assert.NoError(t, pendingJob.Dispatch())
	assert.True(t, fake.AssertNothingPushed())
}

func TestFake_ApplicationBuiltByHand(t *testing.T) {
	app := &Application{}

	assert.NotNil(t, app.Fake())
	assert.Same(t, app.fake, app.Fake())
}

func TestFake_CallingFakeAgainKeepsTheRecords(t *testing.T) {
	app := newFakeTestApp(t)
	fake := app.Fake()
	assert.NoError(t, app.Job(&TestJobOne{}).Dispatch())

	assert.Same(t, fake, app.Fake(&TestJobOne{}))
	assert.NoError(t, app.Job(&TestJobOne{}).Dispatch())

	assert.True(t, fake.AssertPushedTimes(&TestJobOne{}, 2))
}

func TestFake_Reset(t *testing.T) {
	mockDriverCreator := mocksqueue.NewDriverCreator(t)
	mockDriver := mocksqueue.NewDriver(t)
	fake := NewFake()
	fake.enable(nil)

	pendingJob := &PendingJob{
		connection:    "redis",
		driverCreator: mockDriverCreator,
		fake:          fake,
		queue:         "default",
		task:          contractsqueue.Task{ChainJob: contractsqueue.ChainJob{Job: &TestJobOne{}}},
	}
	assert.NoError(t, pendingJob.Dispatch())

	fake.Reset()
	assert.True(t, fake.AssertNothingPushed())

	mockDriverCreator.EXPECT().Create("redis").Return(mockDriver, nil).Once()
	mockDriver.EXPECT().Push(pendingJob.task, "default").Return(nil).Once()
	assert.NoError(t, pendingJob.Dispatch())
	assert.True(t, fake.AssertNothingPushed())
}

func TestFake_RecordsAreCopies(t *testing.T) {
	app := newFakeTestApp(t)
	fake := app.Fake()

	args := []contractsqueue.Arg{{Type: "string", Value: "goravel"}}
	chain := []contractsqueue.ChainJob{{Job: &TestJobOne{}}, {Job: &TestJobTwo{}, Args: []contractsqueue.Arg{{Value: "chained"}}}}
	assert.NoError(t, app.Job(&TestJobOne{}, args).Dispatch())
	assert.NoError(t, app.Chain(chain).Dispatch())
	args[0].Value = "changed"
	chain[1].Args[0].Value = "changed"

	records := fake.Pushed(&TestJobOne{})
	records[0].Args()[0].Value = "changed"
	records[1].Chain()[0].Args[0].Value = "changed"

	records = fake.Pushed(&TestJobOne{})
	assert.Equal(t, "goravel", records[0].Args()[0].Value)
	assert.Equal(t, "chained", records[1].Chain()[0].Args[0].Value)
}

func TestFake_FiltersMayCallTheFake(t *testing.T) {
	app := newFakeTestApp(t)
	fake := app.Fake()
	assert.NoError(t, app.Job(&TestJobOne{}).Dispatch())

	assert.True(t, fake.AssertPushed(&TestJobOne{}, func(contractsqueue.Pushed) bool {
		return fake.AssertPushedTimes(&TestJobOne{}, 1)
	}))
}

func TestFake_Concurrent(t *testing.T) {
	app := newFakeTestApp(t)
	fake := app.Fake()

	var wg sync.WaitGroup
	for range 50 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			assert.NoError(t, app.Job(&TestJobOne{}).Dispatch())
		}()
		go func() {
			defer wg.Done()
			fake.Pushed(&TestJobOne{})
		}()
	}
	wg.Wait()

	assert.True(t, fake.AssertPushedTimes(&TestJobOne{}, 50))
}

func TestPushed_HasChainWithNilJobs(t *testing.T) {
	record := &pushed{chain: []contractsqueue.ChainJob{{}}}

	assert.False(t, record.HasChain(&TestJobOne{}))
	assert.False(t, record.HasChain(nil))
	assert.False(t, record.HasChain())
}
