package queue

const (
	DriverSync   string = "sync"
	DriverCustom string = "custom"
)

type Driver interface {
	// Connection returns the connection name for the driver.
	Connection() string
	// Driver returns the driver name for the driver.
	Driver() string
	// Name returns the name of the driver.
	Name() string
	// Pop pops the next job off of the queue.
	Pop(queue string) (Task, error)
	// Push pushes the job onto the queue.
	Push(task Task, queue string) error
}
