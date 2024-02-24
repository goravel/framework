package session

type Manager interface {
	// BuildSession constructs a new session with the given handler and session ID.
	BuildSession(handler Handler, id ...string) Session
	// Driver retrieves the session driver by name.
	Driver(name ...string) (Handler, error)
	// Extend extends the session manager with a custom driver.
	Extend(driver string, handler func() Handler) Manager
	// Store retrieves a Session by session ID.
	Store(sessionId ...string) Session
}
