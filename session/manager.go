package session

import (
	"fmt"
	"sync"
	"time"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/foundation"
	contractssession "github.com/goravel/framework/contracts/session"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/session/driver"
	"github.com/goravel/framework/support/color"
)

// Ensure interface implementation
var _ contractssession.Manager = (*Manager)(nil)

type Manager struct {
	config      config.Config
	drivers     map[string]contractssession.Driver
	factories   map[string]func() contractssession.Driver
	json        foundation.Json
	sessionPool sync.Pool
	mu          sync.RWMutex
}

func NewManager(config config.Config, json foundation.Json) *Manager {
	manager := &Manager{
		config:    config,
		drivers:   make(map[string]contractssession.Driver),
		factories: make(map[string]func() contractssession.Driver),
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

func (m *Manager) BuildSession(handler contractssession.Driver, sessionID ...string) (contractssession.Session, error) {
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

// Driver retrieves the session driver factory by name and instantiates if needed.
func (m *Manager) Driver(name ...string) (contractssession.Driver, error) {
	driverName := m.getDefaultDriver()
	if len(name) > 0 && name[0] != "" {
		driverName = name[0]
	}

	if driverName == "" {
		return nil, errors.SessionDriverIsNotSet
	}

	m.mu.RLock()
	driverInstance, instanceExists := m.drivers[driverName]
	m.mu.RUnlock()
	if instanceExists {
		return driverInstance, nil
	}

	m.mu.RLock()
	factory, factoryExists := m.factories[driverName]
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	driverInstance, instanceExists = m.drivers[driverName]
	if instanceExists {
		return driverInstance, nil
	}

	if factoryExists {
		driverInstance = factory() // Call the factory registered via Extend/config
		if driverInstance == nil {
			return nil, errors.New(fmt.Sprintf("Factory for session driver '%s' returned nil", driverName))
		}

		m.drivers[driverName] = driverInstance
		m.startGcTimer(driverInstance) // Pass the concrete instance
		return driverInstance, nil
	}

	if driverName == "file" {
		lifetime := m.config.GetInt("session.lifetime")
		fileDriver := driver.NewFile(m.config.GetString("session.files"), lifetime)
		driverInstance = fileDriver
		m.drivers[driverName] = driverInstance
		m.startGcTimer(driverInstance)
		return driverInstance, nil
	}

	return nil, errors.SessionDriverNotSupported.Args(driverName)
}

// Extend registers a factory function for a given driver name.
func (m *Manager) Extend(driver string, factory func() contractssession.Driver) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.factories[driver]; exists {
		return errors.SessionDriverAlreadyExists.Args(driver)
	}
	m.factories[driver] = factory
	return nil
}

func (m *Manager) ReleaseSession(session contractssession.Session) {
	session.Flush().
		SetDriver(nil).
		SetName("").
		SetID("")
	m.sessionPool.Put(session)
}

func (m *Manager) acquireSession() contractssession.Session {
	session := m.sessionPool.Get().(contractssession.Session)
	return session
}

// getDefaultDriver reads the default driver name from config.
func (m *Manager) getDefaultDriver() string {
	return m.config.GetString("session.driver")
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

		// Factory from config: func() (contractssession.Driver, error)
		viaFactoryWithError, ok := viaFactoryAny.(func() (contractssession.Driver, error))
		if !ok {
			continue
		}

		// Wrapper for Extend signature: func() contractssession.Driver
		factoryWrapper := func() contractssession.Driver {
			driverInstance, err := viaFactoryWithError()
			if err != nil {
				color.Errorf("Error creating driver instance for '%s': %s\n", name, err)
				return nil
			}
			if driverInstance == nil {
				color.Errorf("Driver instance for '%s' is nil\n", name)
				return nil
			}
			return driverInstance
		}

		err := m.Extend(name, factoryWrapper)
		if err != nil {
			color.Errorf("Failed to register driver '%s': %s\n", name, err)
		}
	}
}

// startGcTimer remains (it operates on the Driver interface).
func (m *Manager) startGcTimer(driverInstance contractssession.Driver) {
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
