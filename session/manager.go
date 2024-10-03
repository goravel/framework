package session

import (
	"sync"
	"time"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/foundation"
	sessioncontract "github.com/goravel/framework/contracts/session"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/session/driver"
	"github.com/goravel/framework/support/color"
)

type Manager struct {
	config      config.Config
	drivers     map[string]sessioncontract.Driver
	json        foundation.Json
	sessionPool sync.Pool
}

func NewManager(config config.Config, json foundation.Json) *Manager {
	manager := &Manager{
		config:  config,
		drivers: make(map[string]sessioncontract.Driver),
		json:    json,
		sessionPool: sync.Pool{New: func() any {
			return NewSession("", nil, json)
		},
		},
	}
	manager.extendDefaultDrivers()
	return manager
}

func (m *Manager) BuildSession(handler sessioncontract.Driver, sessionID ...string) (sessioncontract.Session, error) {
	if handler == nil {
		return nil, errors.ErrSessionDriverIsNotSet
	}

	session := m.acquireSession()
	session.SetDriver(handler).
		SetName(m.config.GetString("session.cookie"))

	if len(sessionID) > 0 {
		session.SetID(sessionID[0])
	} else {
		session.SetID("")
	}

	return session, nil
}

func (m *Manager) Driver(name ...string) (sessioncontract.Driver, error) {
	var driverName string
	if len(name) > 0 {
		driverName = name[0]
	} else {
		driverName = m.getDefaultDriver()
	}

	if driverName == "" {
		return nil, errors.ErrSessionDriverIsNotSet
	}

	if m.drivers[driverName] == nil {
		return nil, errors.ErrSessionDriverNotSupported.Args(driverName)
	}

	return m.drivers[driverName], nil
}

func (m *Manager) Extend(driver string, handler func() sessioncontract.Driver) error {
	if m.drivers[driver] != nil {
		return errors.ErrSessionDriverAlreadyExists.Args(driver)
	}
	m.drivers[driver] = handler()
	m.startGcTimer(m.drivers[driver])
	return nil
}

func (m *Manager) ReleaseSession(session sessioncontract.Session) {
	session.Flush().
		SetDriver(nil).
		SetName("").
		SetID("")
	m.sessionPool.Put(session)
}

func (m *Manager) acquireSession() sessioncontract.Session {
	session := m.sessionPool.Get().(sessioncontract.Session)
	return session
}

func (m *Manager) getDefaultDriver() string {
	return m.config.GetString("session.driver")
}

func (m *Manager) extendDefaultDrivers() {
	if err := m.Extend("file", m.createFileDriver); err != nil {
		panic(errors.ErrSessionDriverExtensionFailed.Args("file", err))
	}
}

func (m *Manager) createFileDriver() sessioncontract.Driver {
	lifetime := m.config.GetInt("session.lifetime")
	return driver.NewFile(m.config.GetString("session.files"), lifetime)
}

// startGcTimer starts a garbage collection timer for the session driver.
func (m *Manager) startGcTimer(driver sessioncontract.Driver) {
	interval := m.config.GetInt("session.gc_interval", 30)
	if interval <= 0 {
		// No need to start the timer if the interval is zero or negative
		return
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Minute)

	go func() {
		for range ticker.C {
			lifetime := ConfigFacade.GetInt("session.lifetime") * 60
			if err := driver.Gc(lifetime); err != nil {
				color.Red().Printf("Error performing garbage collection: %s\n", err)
			}
		}
	}()
}
