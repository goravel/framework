package queue

type Driver interface {
	// ConnectionName returns the connection name for the driver.
	ConnectionName() string
	// Push pushes the job onto the queue.
	Push(job Job, args []Arg) error
	// Bulk pushes a slice of jobs onto the queue.
	Bulk(jobs []Jobs) error
	// Later pushes the job onto the queue after a delay.
	Later(job Job, delay int) error
	// Pop pops the next job off of the queue.
	Pop() (Job, error)
	// Delete removes a job from the queue.
	Delete(job Job) error
	// Release releases a reserved job back onto the queue.
	Release(job Job, delay int) error
	// Clear clears all pending jobs in the queue.
	Clear() error
	// Size returns the size of the queue.
	Size() (int, error)
	// Server starts a queue server.
	Server(concurrent int)
}
