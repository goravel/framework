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
	json        foundation.Json
	sessionPool sync.Pool
	mu          sync.RWMutex
}

func NewManager(config config.Config, json foundation.Json) *Manager {
	manager := &Manager{
		config:  config,
		drivers: make(map[string]contractssession.Driver),
		json:    json,
		sessionPool: sync.Pool{New: func() any {
			return NewSession("", nil, json)
		},
		},
	}

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

	if driverName == "file" {
		driverInstance = m.file()
		m.startGcTimer(driverInstance)
		return driverInstance, nil
	}

	return nil, errors.SessionDriverNotSupported.Args(driverName)
}

func (m *Manager) Extend(driver string, handler func() contractssession.Driver) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.drivers[driver]; exists {
		return errors.SessionDriverAlreadyExists.Args(driver)
	}

	m.drivers[driver] = handler()
	m.startGcTimer(m.drivers[driver])
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

func (m *Manager) custom(driver string) (contractssession.Driver, error) {
	if custom, ok := m.config.Get(fmt.Sprintf("session.drivers.%s.via", driver)).(contractssession.Driver); ok {
		return custom, nil
	}
	if custom, ok := m.config.Get(fmt.Sprintf("session.drivers.%s.via", driver)).(func() (contractssession.Driver, error)); ok {
		return custom()
	}

	return nil, errors.CacheStoreContractNotFulfilled.Args(driver)
}

func (m *Manager) file() contractssession.Driver {
	lifetime := m.config.GetInt("session.lifetime")
	return driver.NewFile(m.config.GetString("session.files"), lifetime)
}

func (m *Manager) getDefaultDriver() string {
	return m.config.GetString("session.driver")
}

func (m *Manager) registerConfiguredDrivers() error {
	configuredDrivers := m.config.Get("session.drivers", map[string]any{})
	driversMap, ok := configuredDrivers.(map[string]any)
	if !ok {
		return nil
	}
	for name := range driversMap {

		driver := m.config.GetString(fmt.Sprintf("session.drivers.%s.driver", name))

		switch driver {
		case "custom":
			driverInstance, err := m.custom(name)
			if err != nil {
				return err
			}
			m.drivers[name] = driverInstance
			m.startGcTimer(driverInstance)
		default:
			return errors.CacheDriverNotSupported.Args(driver)
		}
	}

	return nil
}

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
