package session

type Manager interface {
	// BuildSession constructs a new session with the given handler and session ID.
	BuildSession(handler Driver, sessionID ...string) Session
	// Driver retrieves the session driver by name.
	Driver(name ...string) (Driver, error)
	// Extend extends the session manager with a custom driver.
	Extend(driver string, handler func() Driver) Manager
}
