package session

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/foundation"
	sessioncontract "github.com/goravel/framework/contracts/session"
	"github.com/goravel/framework/foundation/json"
	mockconfig "github.com/goravel/framework/mocks/config"
)

type ManagerTestSuite struct {
	suite.Suite
	mockConfig *mockconfig.Config
	manager    *Manager
	json       foundation.Json
}

func TestManagerTestSuite(t *testing.T) {
	suite.Run(t, &ManagerTestSuite{})
}

func (m *ManagerTestSuite) SetupTest() {
	m.mockConfig = mockconfig.NewConfig(m.T())
	m.manager = m.getManager()
	m.json = json.NewJson()
}

func (m *ManagerTestSuite) TestDriver() {
	m.mockConfig.On("GetInt", "session.lifetime").Return(120).Once()
	m.mockConfig.On("GetString", "session.files").Return("storage/framework/sessions").Once()

	// provide driver name
	driver, err := m.manager.Driver("file")
	m.Nil(err)
	m.NotNil(driver)
	m.Equal("*driver.File", fmt.Sprintf("%T", driver))

	// provide no driver name
	m.mockConfig.On("GetString", "session.driver").Return("file").Once()
	m.mockConfig.On("GetInt", "session.lifetime").Return(120).Once()
	m.mockConfig.On("GetString", "session.files").Return("storage/framework/sessions").Once()

	driver, err = m.manager.Driver()
	m.Nil(err)
	m.NotNil(driver)
	m.Equal("*driver.File", fmt.Sprintf("%T", driver))

	// provide custom driver
	m.manager.Extend("test", func() sessioncontract.Driver {
		return NewCustomDriver()
	})
	driver, err = m.manager.Driver("test")
	m.Nil(err)
	m.NotNil(driver)
	m.Equal("*session.CustomDriver", fmt.Sprintf("%T", driver))

	// not supported a driver
	m.mockConfig.On("GetString", "session.driver").Return("not_supported").Once()
	driver, err = m.manager.Driver()
	m.NotNil(err)
	m.Equal("driver [not_supported] not supported", err.Error())
	m.Nil(driver)

	// driver is not set
	m.mockConfig.On("GetString", "session.driver").Return("").Once()
	driver, err = m.manager.Driver()
	m.NotNil(err)
	m.Equal("driver is not set", err.Error())
	m.Nil(driver)
}

func (m *ManagerTestSuite) TestExtend() {
	m.manager.Extend("test", func() sessioncontract.Driver {
		return NewCustomDriver()
	})
	driver, err := m.manager.Driver("test")
	m.Nil(err)
	m.NotNil(driver)
	m.Equal("*session.CustomDriver", fmt.Sprintf("%T", driver))
}

func (m *ManagerTestSuite) TestBuildSession() {
	m.mockConfig.On("GetString", "session.cookie").Return("test_cookie").Once()
	session := m.manager.BuildSession(nil)
	m.NotNil(session)
	m.Equal("test_cookie", session.GetName())
}

func (m *ManagerTestSuite) TestGetDefaultDriver() {
	m.mockConfig.On("GetString", "session.driver").Return("file")
	m.Equal("file", m.manager.getDefaultDriver())
}

func (m *ManagerTestSuite) getManager() *Manager {
	return NewManager(m.mockConfig, m.json)
}

type CustomDriver struct{}

func NewCustomDriver() sessioncontract.Driver {
	return &CustomDriver{}
}

func (c *CustomDriver) Close() error {
	return nil
}

func (c *CustomDriver) Destroy(string) error {
	return nil
}

func (c *CustomDriver) Gc(int) error {
	return nil
}

func (c *CustomDriver) Get(string) string {
	return ""
}

func (c *CustomDriver) Open(string, string) error {
	return nil
}

func (c *CustomDriver) Read(string) (string, error) {
	return "", nil
}

func (c *CustomDriver) Write(string, string) error {
	return nil
}
