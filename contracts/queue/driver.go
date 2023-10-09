package queue

type Driver interface {
	// ConnectionName returns the connection name for the driver.
	ConnectionName() string
	// Push pushes the job onto the queue.
	Push(job Job, args []Arg, queue string) error
	// Bulk pushes a slice of jobs onto the queue.
	Bulk(jobs []Jobs, queue string) error
	// Later pushes the job onto the queue after a delay.
	Later(delay int, job Job, args []Arg, queue string) error
	// Pop pops the next job off of the queue.
	Pop(queue string) (Job, []Arg, error)
	// Delete removes a job from the queue.
	Delete(queue string, job Job) error
	// Release releases a reserved job back onto the queue.
	Release(queue string, job Job, delay int) error
	// Clear clears all pending jobs in the queue.
	Clear(queue string) error
	// Size returns the size of the queue.
	Size(queue string) (int64, error)
	// Server starts a queue server.
	Server(concurrent int, queue string)
}
