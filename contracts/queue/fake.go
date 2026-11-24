package queue

import "time"

// Fake records the jobs dispatched while faking instead of pushing them to
// their connection.
//
// The job parameter of the methods below is the job to look for: an instance
// of it, its signature, or a wildcard pattern such as "emails.*" that matches
// every signature sharing the prefix.
type Fake interface {
	// AssertNothingPushed reports whether no faked job was pushed.
	AssertNothingPushed() bool
	// AssertNotPushed reports whether the job was not pushed, and when
	// assertions are given, whether none of its pushes matches all of them.
	AssertNotPushed(job any, assertion ...func(pushed Pushed) bool) bool
	// AssertPushed reports whether the job was pushed, and when assertions
	// are given, whether one of its pushes matches all of them.
	AssertPushed(job any, assertion ...func(pushed Pushed) bool) bool
	// AssertPushedOn reports whether the job was pushed onto the queue, and
	// when assertions are given, whether one of those pushes matches all of
	// them.
	AssertPushedOn(queue string, job any, assertion ...func(pushed Pushed) bool) bool
	// AssertPushedTimes reports whether the job was pushed the given number
	// of times.
	AssertPushedTimes(job any, times int) bool
	// Pushed returns the pushes of the job, only those matching every given
	// filter.
	Pushed(job any, filter ...func(pushed Pushed) bool) []Pushed
	// Reset stops faking: the jobs are pushed again and the records are
	// cleared.
	Reset()
}

// Pushed is a job recorded by the fake, as it would have been pushed.
type Pushed interface {
	// Args returns a copy of the payload of the job. Only the slice is
	// copied: a value of reference type, such as a slice or a map, is shared
	// with the caller.
	Args() []Arg
	// Chain returns a copy of the jobs chained after Job, run one by one
	// after it.
	Chain() []ChainJob
	// Connection returns the connection the job would have been pushed to.
	Connection() string
	// Delay returns when the job would have become available, zero for none.
	Delay() time.Time
	// HasChain reports whether the jobs chained after Job are exactly the
	// given ones, in order. A job is an instance or its signature, compared
	// by signature.
	HasChain(jobs ...any) bool
	// Job returns the job passed to Queue.Job, or the first job passed to
	// Queue.Chain.
	Job() Job
	// Queue returns the queue the job would have been pushed onto.
	Queue() string
}
