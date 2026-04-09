package queue

import "context"

const (
	DriverSync     string = "sync"
	DriverDatabase string = "database"
	DriverCustom   string = "custom"
)

type DriverCreator interface {
	Create(connection string) (Driver, error)
}

type Driver interface {
	// Driver returns the driver name for the driver.
	Driver() string
	// Pop pops the next job off of the queue.
	Pop(queue string) (ReservedJob, error)
	// Push pushes the job onto the queue.
	Push(task Task, queue string) error
}

// DriverWithReceive is an optional interface for drivers that support
// batch message receiving with blocking semantics (e.g., Kafka).
// When a driver implements this interface, the Worker uses Receive
// instead of Pop for message consumption.
type DriverWithReceive interface {
	// Receive retrieves up to count messages from the queue.
	// It may block until at least one message is available or ctx expires.
	// Returns empty slice and nil error if no messages available within deadline.
	Receive(ctx context.Context, queue string, count int) ([]ReservedJob, error)
}
