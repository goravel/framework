package queue

type Job interface {
	Signature() string
	Handle(args ...any) error
}

type Jobs struct {
	Job  Job
	Args []Arg
}
