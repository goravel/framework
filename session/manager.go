package session

import (
	"fmt"

	"github.com/goravel/framework/contracts/config"
	sessioncontract "github.com/goravel/framework/contracts/session"
	"github.com/goravel/framework/session/driver"
)

type Manager struct {
	config        config.Config
	customDrivers map[string]func() sessioncontract.Driver
	drivers       map[string]func() sessioncontract.Driver
}

func NewManager(config config.Config) *Manager {
	manager := &Manager{
		config:        config,
		customDrivers: make(map[string]func() sessioncontract.Driver),
		drivers:       make(map[string]func() sessioncontract.Driver),
	}
	manager.registerDrivers()
	return manager
}

func (m *Manager) BuildSession(handler sessioncontract.Driver, sessionID ...string) sessioncontract.Session {
	return NewSession(m.config.GetString("session.cookie"), handler, sessionID...)
}

func (m *Manager) Driver(name ...string) (sessioncontract.Driver, error) {
	var d string
	if len(name) > 0 {
		d = name[0]
	} else {
		d = m.getDefaultDriver()
	}

	if m.drivers[d] == nil {
		newDriver, err := m.creatDriver(d)
		if err != nil {
			return nil, err
		}

		m.drivers[d] = newDriver
	}

	return m.drivers[d](), nil
}

func (m *Manager) Extend(driver string, handler func() sessioncontract.Driver) sessioncontract.Manager {
	m.customDrivers[driver] = handler
	return m
}

func (m *Manager) Store(sessionId ...string) sessioncontract.Session {
	d, err := m.Driver()
	if err != nil {
		return nil
	}
	return m.BuildSession(d, sessionId...)
}

func (m *Manager) getDefaultDriver() string {
	return m.config.GetString("session.driver")
}

func (m *Manager) creatDriver(name string) (func() sessioncontract.Driver, error) {
	if m.customDrivers[name] != nil {
		return m.customDrivers[name], nil
	}

	if m.drivers[name] != nil {
		return m.drivers[name], nil
	}

	return nil, fmt.Errorf("driver [%s] not supported", name)
}

func (m *Manager) createFileDriver() sessioncontract.Driver {
	lifetime := m.config.GetInt("session.lifetime")
	return driver.NewFileDriver(m.config.GetString("session.files"), lifetime)
}

func (m *Manager) registerDrivers() {
	m.drivers["file"] = m.createFileDriver
}
