package queue

import (
	"reflect"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/support/str"
)

var _ contractsqueue.Fake = (*Fake)(nil)

// Fake records the jobs dispatched while faking. The Queue binding is a
// singleton, so the fake belongs to the Application and every PendingJob it
// creates sees it.
type Fake struct {
	mu       sync.RWMutex
	enabled  atomic.Bool
	all      bool
	patterns []string
	pushed   []*pushed
}

func NewFake() *Fake {
	return &Fake{}
}

func (r *Fake) AssertNothingPushed() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.pushed) == 0
}

func (r *Fake) AssertNotPushed(job any, assertion ...func(pushed contractsqueue.Pushed) bool) bool {
	return !r.AssertPushed(job, assertion...)
}

func (r *Fake) AssertPushed(job any, assertion ...func(pushed contractsqueue.Pushed) bool) bool {
	return len(r.Pushed(job, assertion...)) > 0
}

func (r *Fake) AssertPushedOn(queue string, job any, assertion ...func(pushed contractsqueue.Pushed) bool) bool {
	onQueue := func(pushed contractsqueue.Pushed) bool {
		return pushed.Queue() == queue
	}

	return r.AssertPushed(job, append([]func(pushed contractsqueue.Pushed) bool{onQueue}, assertion...)...)
}

func (r *Fake) AssertPushedTimes(job any, times int) bool {
	return len(r.Pushed(job)) == times
}

func (r *Fake) Pushed(job any, filter ...func(pushed contractsqueue.Pushed) bool) []contractsqueue.Pushed {
	pattern, ok := jobPattern(job)
	if !ok {
		return nil
	}

	// The records are never changed once kept, so they can be handed out
	// as they are.
	var records []contractsqueue.Pushed
	r.mu.RLock()
	for _, record := range r.pushed {
		if matchSignature(pattern, record.signature) {
			records = append(records, record)
		}
	}
	r.mu.RUnlock()

	// Filters run without the lock, so they may call the fake again.
	return slices.DeleteFunc(records, func(record contractsqueue.Pushed) bool {
		return !matchesAll(record, filter)
	})
}

func (r *Fake) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.enabled.Store(false)
	r.all = false
	r.patterns = nil
	r.pushed = nil
}

// enable starts faking the given jobs, every job when there is none. A job
// that has no signature can never be pushed, so it fakes nothing.
func (r *Fake) enable(jobs []any) {
	patterns := make([]string, 0, len(jobs))
	for _, job := range jobs {
		if pattern, ok := jobPattern(job); ok {
			patterns = append(patterns, pattern)
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.enabled.Store(true)
	r.all = len(jobs) == 0
	r.patterns = patterns
}

// record keeps the task when its job is faked, and reports whether it did, in
// which case the task must not be pushed. It takes no lock while nothing is
// faked, so dispatching outside of tests stays lock free.
func (r *Fake) record(task contractsqueue.Task, delay time.Time, connection, queue string) bool {
	if r == nil || !r.enabled.Load() || task.Job == nil {
		return false
	}

	signature := task.Job.Signature()

	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.fakes(signature) {
		return false
	}

	r.pushed = append(r.pushed, &pushed{
		job:        task.Job,
		signature:  signature,
		args:       slices.Clone(task.Args),
		connection: connection,
		queue:      queue,
		delay:      delay,
		chain:      cloneChain(task.Chain),
	})

	return true
}

func (r *Fake) fakes(signature string) bool {
	if !r.enabled.Load() {
		return false
	}
	if r.all {
		return true
	}

	return slices.ContainsFunc(r.patterns, func(pattern string) bool {
		return matchSignature(pattern, signature)
	})
}

// jobPattern returns what a job is looked up by: the signature of an
// instance, or the string itself, a signature or a wildcard pattern.
func jobPattern(job any) (string, bool) {
	var pattern string
	switch j := job.(type) {
	case string:
		pattern = j
	case contractsqueue.Job:
		if isNil(j) {
			return "", false
		}
		pattern = j.Signature()
	}

	return pattern, pattern != ""
}

func matchSignature(pattern, signature string) bool {
	return str.Of(signature).Is(pattern)
}

func matchesAll(pushed contractsqueue.Pushed, filters []func(pushed contractsqueue.Pushed) bool) bool {
	for _, filter := range filters {
		if filter != nil && !filter(pushed) {
			return false
		}
	}

	return true
}

func cloneChain(chain []contractsqueue.ChainJob) []contractsqueue.ChainJob {
	chain = slices.Clone(chain)
	for i := range chain {
		chain[i].Args = slices.Clone(chain[i].Args)
	}

	return chain
}

// isNil reports whether the value is nil, a nil pointer held in an interface
// included.
func isNil(value any) bool {
	if value == nil {
		return true
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}

var _ contractsqueue.Pushed = (*pushed)(nil)

// pushed is a record of the fake. It is never changed once kept, and the
// slices are handed out as copies, so a test can not change it.
type pushed struct {
	job        contractsqueue.Job
	signature  string
	args       []contractsqueue.Arg
	connection string
	queue      string
	delay      time.Time
	chain      []contractsqueue.ChainJob
}

func (r *pushed) Args() []contractsqueue.Arg {
	return slices.Clone(r.args)
}

func (r *pushed) Chain() []contractsqueue.ChainJob {
	return cloneChain(r.chain)
}

func (r *pushed) Connection() string {
	return r.connection
}

func (r *pushed) Delay() time.Time {
	return r.delay
}

func (r *pushed) HasChain(jobs ...any) bool {
	if len(r.chain) != len(jobs) {
		return false
	}

	for i, job := range jobs {
		signature, ok := jobPattern(job)
		if !ok || isNil(r.chain[i].Job) || r.chain[i].Job.Signature() != signature {
			return false
		}
	}

	return true
}

func (r *pushed) Job() contractsqueue.Job {
	return r.job
}

func (r *pushed) Queue() string {
	return r.queue
}
