package session

type Manager interface {
	Driver(name ...string) (Handler, error)
	Extend(driver string, handler func() Handler) Manager
	Store(sessionId ...string) Session
}
