package queue

import (
	"time"
)

//go:generate mockery --name=Task
type Task interface {
	Dispatch() error
	DispatchSync() error
	Delay(time time.Time) Task
	OnConnection(connection string) Task
	OnQueue(queue string) Task
}
