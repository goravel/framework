package queue

import "time"

const (
	DriverSync   string = "sync"
	DriverCustom string = "custom"
)

type Driver interface {
	// Connection returns the connection name for the driver.
	Connection() string
	// Driver returns the driver name for the driver.
	Driver() string
	// Later pushes the job onto the queue after a delay.
	Later(delay time.Time, task Task, queue string) error
	// Name returns the name of the driver.
	Name() string
	// Pop pops the next job off of the queue.
	Pop(queue string) (*Task, error)
	// Push pushes the job onto the queue.
	Push(task Task, queue string) error
}
