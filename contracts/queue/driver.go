package queue

type Driver interface {
	// ConnectionName returns the connection name for the driver.
	ConnectionName() string
	// DriverName returns the driver name for the driver.
	DriverName() string
	// Push pushes the job onto the queue.
	Push(job Job, args []Arg, queue string) error
	// Bulk pushes a slice of jobs onto the queue.
	Bulk(jobs []Jobs, queue string) error
	// Later pushes the job onto the queue after a delay.
	Later(delay uint, job Job, args []Arg, queue string) error
	// Pop pops the next job off of the queue.
	Pop(queue string) (Job, []Arg, error)
	// Delete removes a job from the queue.
	Delete(queue string, job Jobs) error
	// Release releases a reserved job back onto the queue.
	Release(queue string, job Jobs, delay uint) error
	// Clear clears all pending jobs in the queue.
	Clear(queue string) error
	// Size returns the size of the queue.
	Size(queue string) (uint64, error)
}
