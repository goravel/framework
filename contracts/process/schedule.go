package process

import "time"

type Priority int

const (
	PriorityLow      Priority = 100
	PriorityNormal   Priority = 200 // Default priority
	PriorityHigh     Priority = 300
	PriorityCritical Priority = 400
)

// Schedulable defines the contract for a command that can be scheduled within the process pool.
//
// It provides a stable interface for different scheduling strategies to query
// the properties of a command (like its priority and timeout) in order to make
// an intelligent decision about the optimal execution order. This interface is
// implemented by types such as PoolCommand.
type Schedulable interface {
	// GetKey returns the unique string key assigned to the command. This key
	// is used to identify the process in the final results map and in any
	// real-time output handlers.
	GetKey() string

	// GetTimeout returns the configured maximum execution duration for the command.
	// This allows a strategy to prioritize tasks with shorter or more urgent deadlines.
	// A time.Duration of zero should be interpreted by strategies as an infinite
	// timeout, effectively giving it the lowest possible priority in a timeout-based sort.
	GetTimeout() time.Duration

	// GetPriority returns the configured priority level for the command.
	// Strategies use this to ensure that more important tasks are executed before
	// less important ones, regardless of their timeout. A higher numerical value
	// typically indicates a higher execution priority.
	GetPriority() Priority
}

// Strategy defines an interface for process pool scheduling algorithms.
// It takes a slice of schedulable commands and returns a re-ordered slice
// based on its internal logic before execution begins.
type Strategy interface {
	Schedule([]Schedulable) []Schedulable
}
