package queue

import "time"

// Fake records the jobs dispatched while faking instead of pushing them to
// their connection. Jobs are told apart by their signature.
type Fake interface {
	// AssertNothingPushed reports whether no faked job was pushed.
	AssertNothingPushed() bool
	// AssertNotPushed reports whether the job was not pushed, and when
	// assertions are given, whether none of its pushes matches all of them.
	AssertNotPushed(job Job, assertion ...func(pushed Pushed) bool) bool
	// AssertPushed reports whether the job was pushed, and when assertions
	// are given, whether one of its pushes matches all of them.
	AssertPushed(job Job, assertion ...func(pushed Pushed) bool) bool
	// AssertPushedOn reports whether the job was pushed onto the queue, and
	// when assertions are given, whether one of those pushes matches all of
	// them.
	AssertPushedOn(queue string, job Job, assertion ...func(pushed Pushed) bool) bool
	// AssertPushedTimes reports whether the job was pushed exactly times
	// times.
	AssertPushedTimes(job Job, times int) bool
	// Pushed returns the pushes of the job, only those matching every given
	// filter.
	Pushed(job Job, filter ...func(pushed Pushed) bool) []Pushed
	// Reset stops faking: the jobs are pushed again and the records are
	// cleared.
	Reset()
}

// Pushed is a job recorded by the fake, as it would have been pushed.
type Pushed struct {
	// Job is the job passed to Job, or the first job passed to Chain.
	Job Job
	// Args is the payload of the job.
	Args []Arg
	// Connection is the connection the job would have been pushed to.
	Connection string
	// Queue is the queue the job would have been pushed onto.
	Queue string
	// Delay is when the job would have become available, zero for none.
	Delay time.Time
	// Chain is the rest of the jobs passed to Chain, run one by one after
	// Job.
	Chain []ChainJob
}

// HasChain reports whether the jobs chained after Job are exactly the given
// ones, in order, compared by signature.
func (r Pushed) HasChain(jobs ...Job) bool {
	if len(r.Chain) != len(jobs) {
		return false
	}

	for i, job := range jobs {
		if r.Chain[i].Job == nil || job == nil || r.Chain[i].Job.Signature() != job.Signature() {
			return false
		}
	}

	return true
}
