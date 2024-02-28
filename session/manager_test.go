package session

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	sessioncontract "github.com/goravel/framework/contracts/session"
	mockconfig "github.com/goravel/framework/mocks/config"
	mockfilesystem "github.com/goravel/framework/mocks/filesystem"
	mockfoundation "github.com/goravel/framework/mocks/foundation"
)

type ManagerTestSuite struct {
	suite.Suite
	mockApp     *mockfoundation.Application
	mockConfig  *mockconfig.Config
	mockStorage *mockfilesystem.Storage
}

func TestManagerTestSuite(t *testing.T) {
	suite.Run(t, &ManagerTestSuite{})
}

func (m *ManagerTestSuite) SetupTest() {
	m.mockApp = mockfoundation.NewApplication(m.T())
	m.mockConfig = mockconfig.NewConfig(m.T())
	m.mockStorage = mockfilesystem.NewStorage(m.T())
}

func (m *ManagerTestSuite) TestDriver() {
	manager := m.getManager()
	m.mockConfig.On("GetInt", "session.lifetime").Once().Return(120)
	m.mockConfig.On("GetString", "session.files").Once().Return("storage/framework/sessions")
	// provide driver name
	m.mockApp.On("MakeStorage").Once().Return(m.mockStorage)
	driver, err := manager.Driver("file")
	m.Nil(err)
	m.NotNil(driver)
	m.Equal("*driver.FileDriver", fmt.Sprintf("%T", driver))

	// provide no driver name
	m.mockConfig.On("GetString", "session.driver").Once().Return("file")
	m.mockConfig.On("GetInt", "session.lifetime").Once().Return(120)
	m.mockConfig.On("GetString", "session.files").Once().Return("storage/framework/sessions")

	m.mockApp.On("MakeStorage").Once().Return(m.mockStorage)
	driver, err = manager.Driver()
	m.Nil(err)
	m.NotNil(driver)
	m.Equal("*driver.FileDriver", fmt.Sprintf("%T", driver))

	// provide custom driver
	manager.Extend("test", func() sessioncontract.Driver {
		return NewCustomDriver()
	})
	driver, err = manager.Driver("test")
	m.Nil(err)
	m.NotNil(driver)
	m.Equal("*session.CustomDriver", fmt.Sprintf("%T", driver))

	// not supported a driver
	m.mockConfig.On("GetString", "session.driver").Once().Return("not_supported")
	driver, err = manager.Driver()
	m.NotNil(err)
	m.Equal("driver [not_supported] not supported", err.Error())
	m.Nil(driver)
}

func (m *ManagerTestSuite) TestExtend() {
	manager := m.getManager()
	manager.Extend("test", func() sessioncontract.Driver {
		return NewCustomDriver()
	})
	driver, err := manager.Driver("test")
	m.Nil(err)
	m.NotNil(driver)
	m.Equal("*session.CustomDriver", fmt.Sprintf("%T", driver))
}

func (m *ManagerTestSuite) TestBuildSession() {
	manager := m.getManager()
	m.mockConfig.On("GetString", "session.cookie").Once().Return("test_cookie")
	session := manager.BuildSession(nil)
	m.NotNil(session)
	m.Equal("test_cookie", session.GetName())
}

func (m *ManagerTestSuite) TestGetDefaultDriver() {
	manager := m.getManager()
	m.mockConfig.On("GetString", "session.driver").Return("file")
	m.Equal("file", manager.getDefaultDriver())
}

func (m *ManagerTestSuite) getManager() *Manager {
	m.mockApp.On("MakeConfig").Once().Return(m.mockConfig)
	return NewManager(m.mockApp)
}

type CustomDriver struct{}

func NewCustomDriver() sessioncontract.Driver {
	return &CustomDriver{}
}

func (c *CustomDriver) Close() bool {
	return true
}

func (c *CustomDriver) Destroy(string) error {
	return nil
}

func (c *CustomDriver) Gc(int) int {
	return 0
}

func (c *CustomDriver) Get(string) string {
	return ""
}

func (c *CustomDriver) Open(string, string) bool {
	return true
}

func (c *CustomDriver) Read(string) string {
	return ""
}

func (c *CustomDriver) Write(string, string) error {
	return nil
}
