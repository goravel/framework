package queue

import (
	"slices"
	"sync"

	contractsqueue "github.com/goravel/framework/contracts/queue"
)

var _ contractsqueue.Fake = (*Fake)(nil)

// Fake records the jobs dispatched while faking. The Queue binding is a
// singleton, so the fake belongs to the Application and every PendingJob it
// creates sees it.
type Fake struct {
	mu         sync.RWMutex
	enabled    bool
	all        bool
	signatures []string
	pushed     []contractsqueue.Pushed
}

func NewFake() *Fake {
	return &Fake{}
}

func (r *Fake) AssertNothingPushed() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.pushed) == 0
}

func (r *Fake) AssertNotPushed(job contractsqueue.Job, assertion ...func(pushed contractsqueue.Pushed) bool) bool {
	return !r.AssertPushed(job, assertion...)
}

func (r *Fake) AssertPushed(job contractsqueue.Job, assertion ...func(pushed contractsqueue.Pushed) bool) bool {
	return len(r.Pushed(job, assertion...)) > 0
}

func (r *Fake) AssertPushedOn(queue string, job contractsqueue.Job, assertion ...func(pushed contractsqueue.Pushed) bool) bool {
	onQueue := func(pushed contractsqueue.Pushed) bool {
		return pushed.Queue == queue
	}

	return r.AssertPushed(job, append([]func(pushed contractsqueue.Pushed) bool{onQueue}, assertion...)...)
}

func (r *Fake) AssertPushedTimes(job contractsqueue.Job, times int) bool {
	return len(r.Pushed(job)) == times
}

func (r *Fake) Pushed(job contractsqueue.Job, filter ...func(pushed contractsqueue.Pushed) bool) []contractsqueue.Pushed {
	if job == nil {
		return nil
	}
	signature := job.Signature()

	r.mu.RLock()
	records := slices.Clone(r.pushed)
	r.mu.RUnlock()

	// Filters run without the lock, so they may call the fake again.
	result := make([]contractsqueue.Pushed, 0, len(records))
	for _, record := range records {
		if record.Job.Signature() != signature {
			continue
		}

		record = clonePushed(record)
		if matchesAll(record, filter) {
			result = append(result, record)
		}
	}

	return slices.Clip(result)
}

func (r *Fake) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.enabled = false
	r.all = false
	r.signatures = nil
	r.pushed = nil
}

// enable starts faking the given jobs, every job when there is none.
func (r *Fake) enable(jobs []contractsqueue.Job) {
	signatures := make([]string, 0, len(jobs))
	for _, job := range jobs {
		if job != nil {
			signatures = append(signatures, job.Signature())
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.enabled = true
	r.all = len(jobs) == 0
	r.signatures = signatures
}

// fakes reports whether the job is faked.
func (r *Fake) fakes(job contractsqueue.Job) bool {
	if r == nil || job == nil {
		return false
	}

	signature := job.Signature()

	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.fakesLocked(signature)
}

func (r *Fake) fakesLocked(signature string) bool {
	return r.enabled && (r.all || slices.Contains(r.signatures, signature))
}

// record keeps the task when its job is faked, and reports whether it did,
// in which case the task must not be pushed.
func (r *Fake) record(task contractsqueue.Task, connection, queue string) bool {
	if r == nil || task.Job == nil {
		return false
	}

	signature := task.Job.Signature()

	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.fakesLocked(signature) {
		return false
	}

	r.pushed = append(r.pushed, clonePushed(contractsqueue.Pushed{
		Job:        task.Job,
		Args:       task.Args,
		Connection: connection,
		Queue:      queue,
		Delay:      task.Delay,
		Chain:      task.Chain,
	}))

	return true
}

// clonePushed copies the slices of the record, so the records and the ones
// handed out never share them with the caller.
func clonePushed(pushed contractsqueue.Pushed) contractsqueue.Pushed {
	pushed.Args = slices.Clone(pushed.Args)
	pushed.Chain = slices.Clone(pushed.Chain)
	for i := range pushed.Chain {
		pushed.Chain[i].Args = slices.Clone(pushed.Chain[i].Args)
	}

	return pushed
}

func matchesAll(pushed contractsqueue.Pushed, filters []func(pushed contractsqueue.Pushed) bool) bool {
	for _, filter := range filters {
		if filter != nil && !filter(pushed) {
			return false
		}
	}

	return true
}
