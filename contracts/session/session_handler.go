package session

// Handler is the interface for Session handlers.
type Handler interface {
	// Close closes the session handler.
	Close() bool
	// Destroy destroys the session with the given ID.
	Destroy(id string) bool
	// Gc performs garbage collection on the session handler with the given maximum lifetime.
	Gc(maxLifetime int) int
	// Open opens a session with the given path and name.
	Open(path string, name string) bool
	// Read reads the session data associated with the given ID.
	Read(id string) string
	// Write writes the session data associated with the given ID.
	Write(id string, data string) error
}
