package session

import (
	"fmt"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/foundation"
	sessioncontract "github.com/goravel/framework/contracts/session"
	"github.com/goravel/framework/session/driver"
)

type Manager struct {
	config        config.Config
	customDrivers map[string]sessioncontract.Driver
	drivers       map[string]sessioncontract.Driver
	json          foundation.Json
}

func NewManager(config config.Config, json foundation.Json) *Manager {
	manager := &Manager{
		config:        config,
		customDrivers: make(map[string]sessioncontract.Driver),
		drivers:       make(map[string]sessioncontract.Driver),
		json:          json,
	}
	manager.registerDrivers()
	return manager
}

func (m *Manager) BuildSession(handler sessioncontract.Driver, sessionID ...string) sessioncontract.Session {
	return NewSession(m.config.GetString("session.cookie"), handler, m.json, sessionID...)
}

func (m *Manager) Driver(name ...string) (sessioncontract.Driver, error) {
	var driverName string
	if len(name) > 0 {
		driverName = name[0]
	} else {
		driverName = m.getDefaultDriver()
	}

	if driverName == "" {
		return nil, fmt.Errorf("driver is not set")
	}

	if m.drivers[driverName] == nil {
		if m.customDrivers[driverName] == nil {
			return nil, fmt.Errorf("driver [%s] not supported", driverName)
		}

		m.drivers[driverName] = m.customDrivers[driverName]
	}

	return m.drivers[driverName], nil
}

func (m *Manager) Extend(driver string, handler func() sessioncontract.Driver) sessioncontract.Manager {
	m.customDrivers[driver] = handler()
	return m
}

func (m *Manager) getDefaultDriver() string {
	return m.config.GetString("session.driver")
}

func (m *Manager) createFileDriver() sessioncontract.Driver {
	lifetime := m.config.GetInt("session.lifetime")
	return driver.NewFile(m.config.GetString("session.files"), lifetime)
}

func (m *Manager) registerDrivers() {
	m.drivers["file"] = m.createFileDriver()
}
