package queue

import "time"

const DriverSync string = "sync"
const DriverCustom string = "custom"

type Driver interface {
	// Bulk pushes a slice of jobs onto the queue.
	Bulk(jobs []Jobs, queue string) error
	// Connection returns the connection name for the driver.
	Connection() string
	// Driver returns the driver name for the driver.
	Driver() string
	// Later pushes the job onto the queue after a delay.
	Later(delay time.Time, job Job, args []any, queue string) error
	// Pop pops the next job off of the queue.
	Pop(queue string) (Job, []any, error)
	// Push pushes the job onto the queue.
	Push(job Job, args []any, queue string) error
}
