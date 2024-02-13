package session

import (
	"fmt"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/foundation"
	sessioncontract "github.com/goravel/framework/contracts/session"
)

type Manager struct {
	app            foundation.Application
	config         config.Config
	customCreators map[string]func() sessioncontract.Handler
	drivers        map[string]sessioncontract.Handler
}

func NewManager(app foundation.Application) *Manager {
	con := app.MakeConfig()
	manager := &Manager{
		app:            app,
		config:         con,
		customCreators: make(map[string]func() sessioncontract.Handler),
		drivers:        make(map[string]sessioncontract.Handler),
	}
	manager.registerDrivers()
	return manager
}

func (m *Manager) BuildSession(handler sessioncontract.Handler, sessionId ...string) *Store {
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

	return m.drivers[driver], nil
}

func (m *Manager) Extend(driver string, handler func() sessioncontract.Handler) sessioncontract.Manager {
	m.customCreators[driver] = handler
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

func (m *Manager) callCustomCreator(driver string) sessioncontract.Handler {
	return m.customCreators[driver]()
}

func (m *Manager) creatDriver(name string) (sessioncontract.Handler, error) {
	if m.customCreators[name] != nil {
		return m.customCreators[name](), nil
	}

	if m.drivers[name] != nil {
		return m.drivers[name], nil
	}

	return nil, fmt.Errorf("driver [%s] not supported", name)
}

func (m *Manager) createFileDriver() sessioncontract.Handler {
	lifetime := m.config.GetInt("session.lifetime")
	return NewFileHandler(m.config.GetString("session.files"), lifetime)
}

func (m *Manager) registerDrivers() {
	m.drivers["file"] = m.createFileDriver()
}
