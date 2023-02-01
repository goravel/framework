package event

//go:generate mockery --name=Instance
type Instance interface {
	Register(map[Event][]Listener)
	Job(event Event, args []Arg) Task
	GetEvents() map[Event][]Listener
}

type Event interface {
	Handle(args []Arg) ([]Arg, error)
}

type Listener interface {
	Signature() string
	Queue(args ...any) Queue
	Handle(args ...any) error
}

//go:generate mockery --name=Task
type Task interface {
	Dispatch() error
}

type Arg struct {
	Type  string
	Value any
}

type Queue struct {
	Enable     bool
	Connection string
	Queue      string
}
