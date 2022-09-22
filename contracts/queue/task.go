package queue

//go:generate mockery --name=Task --output=../mocks/queue/ --outpkg=queue --keeptree
type Task interface {
	Dispatch() error
	DispatchSync() error
	OnConnection(connection string) Task
	OnQueue(queue string) Task
}
