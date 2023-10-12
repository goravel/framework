package queue

import (
	"time"
)

//go:generate mockery --name=Task
type Task interface {
	// Dispatch dispatches the task.
	Dispatch() error
	// DispatchSync dispatches the task synchronously.
	DispatchSync() error
	// Delay dispatches the task after the given delay.
	Delay(time time.Time) Task
	// OnConnection sets the connection of the task.
	OnConnection(connection string) Task
	// OnQueue sets the queue of the task.
	OnQueue(queue string) Task
}
