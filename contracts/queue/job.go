package queue

type Job interface {
	Signature() string
	Handle(args ...interface{}) error
}

type Jobs struct {
	Job  Job
	Args []Arg
}
