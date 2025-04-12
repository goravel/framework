package session

import (
	"fmt"
	"sync"
	"time"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/foundation"
	sessioncontract "github.com/goravel/framework/contracts/session"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/color"
)

type Manager struct {
	config      config.Config
	drivers     map[string]sessioncontract.Driver
	factories   map[string]func() sessioncontract.Driver
	json        foundation.Json
	sessionPool sync.Pool
	mu          sync.RWMutex
}

func NewManager(config config.Config, json foundation.Json) *Manager {
	manager := &Manager{
		config:    config,
		drivers:   make(map[string]sessioncontract.Driver),
		factories: make(map[string]func() sessioncontract.Driver),
		json:      json,
		sessionPool: sync.Pool{New: func() any {
			return NewSession("", nil, json)
		},
		},
	}
	// Reads config ONLY to find drivers.
	manager.registerConfiguredDrivers()
	return manager
}

// Driver retrieves the session driver factory by name and instantiates if needed.
func (m *Manager) Driver(name ...string) (sessioncontract.Driver, error) {
	driverName := m.getDefaultDriver()
	if len(name) > 0 && name[0] != "" {
		driverName = name[0]
	}

	if driverName == "" {
		return nil, errors.SessionDriverIsNotSet
	}

	// Check instance cache first
	m.mu.RLock()
	driverInstance, instanceExists := m.drivers[driverName]
	m.mu.RUnlock()
	if instanceExists {
		return driverInstance, nil
	}

	// Resolve factory
	m.mu.RLock()
	factory, factoryExists := m.factories[driverName]
	m.mu.RUnlock()

	if !factoryExists {
		return nil, errors.SessionDriverNotSupported.Args(driverName)
	}

	// Instantiate using factory (with lock)
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check instance existence after acquiring lock
	driverInstance, instanceExists = m.drivers[driverName]
	if instanceExists {
		return driverInstance, nil
	}

	driverInstance = factory()
	if driverInstance == nil {
		return nil, errors.New(fmt.Sprintf("Session driver %s factory returned nil", driverName))
	}

	m.drivers[driverName] = driverInstance
	m.startGcTimer(driverInstance) // Start GC for the new instance

	return driverInstance, nil
}

// Extend registers a factory function for a given driver name.
func (m *Manager) Extend(driver string, factory func() sessioncontract.Driver) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.factories[driver]; exists {
		return errors.SessionDriverAlreadyExists.Args(driver)
	}
	m.factories[driver] = factory
	return nil
}

// registerConfiguredDrivers ONLY reads from config.
func (m *Manager) registerConfiguredDrivers() {
	configuredDrivers := m.config.Get("session.drivers", map[string]any{})
	driversMap, ok := configuredDrivers.(map[string]any)
	if !ok {
		return
	}
	for name, driverConfigAny := range driversMap {
		driverConfig, ok := driverConfigAny.(map[string]any)
		if !ok {

			continue
		}

		viaFactoryAny, exists := driverConfig["via"]
		if !exists {

			continue
		}

		// Factory from config: func() (sessioncontract.Driver, error)
		viaFactoryWithError, ok := viaFactoryAny.(func() (sessioncontract.Driver, error))
		if !ok {

			continue
		}

		// Wrapper for Extend signature: func() sessioncontract.Driver
		factoryWrapper := func() sessioncontract.Driver {
			driverInstance, err := viaFactoryWithError()
			if err != nil {

				return nil
			}
			if driverInstance == nil {

				return nil
			}
			return driverInstance
		}

		// Register using Extend. Errors logged internally by Extend.
		_ = m.Extend(name, factoryWrapper)
	}
}

func (m *Manager) BuildSession(handler sessioncontract.Driver, sessionID ...string) (sessioncontract.Session, error) {
	if handler == nil {
		return nil, errors.SessionDriverIsNotSet
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

// getDefaultDriver reads the default driver name from config.
func (m *Manager) getDefaultDriver() string {
	return m.config.GetString("session.driver")
}

// startGcTimer remains (it operates on the Driver interface).
func (m *Manager) startGcTimer(driverInstance sessioncontract.Driver) {
	interval := m.config.GetInt("session.gc_interval")
	if interval <= 0 {
		// No need to start the timer if the interval is zero or negative
		return
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Minute)

	go func() {
		for range ticker.C {
			lifetime := ConfigFacade.GetInt("session.lifetime") * 60
			if err := driverInstance.Gc(lifetime); err != nil {
				color.Errorf("Error performing garbage collection: %s\n", err)
			}
		}
	}()
}

// Ensure interface implementation
var _ sessioncontract.Manager = (*Manager)(nil)
