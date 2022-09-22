package queue

//go:generate mockery --name=Job --output=../mocks/queue/ --outpkg=queue --keeptree
type Job interface {
	Signature() string
	Handle(args ...interface{}) error
}

type Jobs struct {
	Job  Job
	Args []Arg
}
