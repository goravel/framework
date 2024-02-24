package session

import (
	"fmt"

	"github.com/goravel/framework/contracts/config"
	sessioncontract "github.com/goravel/framework/contracts/session"
	"github.com/goravel/framework/session/handler"
)

type Manager struct {
	config        config.Config
	customDrivers map[string]func() sessioncontract.Handler
	drivers       map[string]func() sessioncontract.Handler
}

func NewManager(config config.Config) *Manager {
	manager := &Manager{
		config:        config,
		customDrivers: make(map[string]func() sessioncontract.Handler),
		drivers:       make(map[string]func() sessioncontract.Handler),
	}
	manager.registerDrivers()
	return manager
}

func (m *Manager) BuildSession(handler sessioncontract.Handler, sessionId ...string) sessioncontract.Session {
	return NewStore(m.config.GetString("session.cookie"), handler, sessionId...)
}

func (m *Manager) Driver(name ...string) (sessioncontract.Handler, error) {
	var driver string
	if len(name) > 0 {
		driver = name[0]
	} else {
		driver = m.getDefaultDriver()
	}

	if m.drivers[driver] == nil {
		newDriver, err := m.creatDriver(driver)
		if err != nil {
			return nil, err
		}

		m.drivers[driver] = newDriver
	}

	return m.drivers[driver](), nil
}

func (m *Manager) Extend(driver string, handler func() sessioncontract.Handler) sessioncontract.Manager {
	m.customDrivers[driver] = handler
	return m
}

func (m *Manager) Store(sessionId ...string) sessioncontract.Session {
	driver, err := m.Driver()
	if err != nil {
		return nil
	}
	return m.BuildSession(driver, sessionId...)
}

func (m *Manager) getDefaultDriver() string {
	return m.config.GetString("session.driver")
}

func (m *Manager) creatDriver(name string) (func() sessioncontract.Handler, error) {
	if m.customDrivers[name] != nil {
		return m.customDrivers[name], nil
	}

	if m.drivers[name] != nil {
		return m.drivers[name], nil
	}

	return nil, fmt.Errorf("driver [%s] not supported", name)
}

func (m *Manager) createFileDriver() sessioncontract.Handler {
	lifetime := m.config.GetInt("session.lifetime")
	return handler.NewFileHandler(m.config.GetString("session.files"), lifetime)
}

func (m *Manager) registerDrivers() {
	m.drivers["file"] = m.createFileDriver
}
