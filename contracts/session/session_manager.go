package session

import "github.com/goravel/framework/contracts/foundation"

type Manager interface {
	Driver(name ...string) (Handler, error)
	Extend(driver string, handler func(app foundation.Application) Handler) Manager
	Store(sessionId ...string) Session
}
