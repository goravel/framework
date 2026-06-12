package console

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/console/command"
	frameworkerrors "github.com/goravel/framework/errors"
	mockscache "github.com/goravel/framework/mocks/cache"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksfilesystem "github.com/goravel/framework/mocks/filesystem"
	mockshash "github.com/goravel/framework/mocks/hash"
	mocksview "github.com/goravel/framework/mocks/view"
)

type DownCommandTestSuite struct {
	suite.Suite
	mockHash    *mockshash.Hash
	mockView    *mocksview.View
	mockCache   *mockscache.Cache
	mockConfig  *mocksconfig.Config
	mockStorage *mocksfilesystem.Storage
}

func TestDownCommandTestSuite(t *testing.T) {
	suite.Run(t, new(DownCommandTestSuite))
}

func (s *DownCommandTestSuite) SetupTest() {
	s.mockHash = mockshash.NewHash(s.T())
	s.mockView = mocksview.NewView(s.T())
	s.mockCache = mockscache.NewCache(s.T())
	s.mockConfig = mocksconfig.NewConfig(s.T())
	s.mockStorage = mocksfilesystem.NewStorage(s.T())
}

func (s *DownCommandTestSuite) maintenance() *MaintenanceMode {
	return NewMaintenanceMode(s.mockConfig, s.mockCache, s.mockStorage)
}

func (s *DownCommandTestSuite) TestSignature() {
	expected := "down"
	s.Require().Equal(expected, NewDownCommand(s.mockView, s.mockHash, s.maintenance()).Signature())
}

func (s *DownCommandTestSuite) TestDescription() {
	expected := "Put the application into maintenance mode"
	s.Require().Equal(expected, NewDownCommand(s.mockView, s.mockHash, s.maintenance()).Description())
}

func (s *DownCommandTestSuite) TestExtend() {
	cmd := NewDownCommand(s.mockView, s.mockHash, s.maintenance())
	got := cmd.Extend()

	s.Equal(6, len(got.Flags))
}

func (s *DownCommandTestSuite) TestHandle() {
	cmd := NewDownCommand(s.mockView, s.mockHash, s.maintenance())

	flag := cmd.Extend().Flags[0].(*command.StringFlag)
	mockCtx := mocksconsole.NewContext(s.T())
	mockCtx.EXPECT().OptionInt("status").Return(503)
	mockCtx.EXPECT().Option("render").Return("")
	mockCtx.EXPECT().Option("redirect").Return("")
	mockCtx.EXPECT().Option("secret").Return("")
	mockCtx.EXPECT().OptionBool("with-secret").Return(false)
	mockCtx.EXPECT().Option("reason").Return(flag.Value)
	mockCtx.EXPECT().Success("The application is in maintenance mode now").Once()
	s.mockConfig.EXPECT().GetString("APP_MAINTENANCE_DRIVER", "file").Return("file").Once()
	s.mockStorage.EXPECT().Put("framework/maintenance.json", "{\"reason\":\"The application is under maintenance\",\"status\":503}").Return(nil)

	err := cmd.Handle(mockCtx)

	assert.Nil(s.T(), err)
}

func (s *DownCommandTestSuite) TestHandleWithReason() {
	mockCtx := mocksconsole.NewContext(s.T())
	mockCtx.EXPECT().OptionInt("status").Return(505)
	mockCtx.EXPECT().Option("reason").Return("Under maintenance").Once()
	mockCtx.EXPECT().Option("redirect").Return("").Once()
	mockCtx.EXPECT().Option("render").Return("").Once()
	mockCtx.EXPECT().Option("secret").Return("").Once()
	mockCtx.EXPECT().OptionBool("with-secret").Return(false).Once()
	mockCtx.EXPECT().Success("The application is in maintenance mode now").Once()
	s.mockConfig.EXPECT().GetString("APP_MAINTENANCE_DRIVER", "file").Return("file").Once()
	s.mockStorage.EXPECT().Put("framework/maintenance.json", "{\"reason\":\"Under maintenance\",\"status\":505}").Return(nil).Once()

	cmd := NewDownCommand(s.mockView, s.mockHash, s.maintenance())
	err := cmd.Handle(mockCtx)

	assert.Nil(s.T(), err)
}

func (s *DownCommandTestSuite) TestHandleWithRedirect() {
	mockCtx := mocksconsole.NewContext(s.T())
	mockCtx.EXPECT().OptionInt("status").Return(503).Once()
	mockCtx.EXPECT().Option("render").Return("").Once()
	mockCtx.EXPECT().Option("redirect").Return("/maintenance").Once()
	mockCtx.EXPECT().Option("secret").Return("").Once()
	mockCtx.EXPECT().OptionBool("with-secret").Return(false).Once()
	mockCtx.EXPECT().Success("The application is in maintenance mode now").Once()
	s.mockConfig.EXPECT().GetString("APP_MAINTENANCE_DRIVER", "file").Return("file").Once()
	s.mockStorage.EXPECT().Put("framework/maintenance.json", "{\"redirect\":\"/maintenance\",\"status\":503}").Return(nil).Once()

	cmd := NewDownCommand(s.mockView, s.mockHash, s.maintenance())
	err := cmd.Handle(mockCtx)

	assert.Nil(s.T(), err)
}

func (s *DownCommandTestSuite) TestHandleWithRender() {
	s.mockView.EXPECT().Exists("errors/503.tmpl").Return(true)

	mockCtx := mocksconsole.NewContext(s.T())
	mockCtx.EXPECT().OptionInt("status").Return(503).Once()
	mockCtx.EXPECT().Option("render").Return("errors/503.tmpl").Once()
	mockCtx.EXPECT().Option("redirect").Return("").Once()
	mockCtx.EXPECT().Option("secret").Return("").Once()
	mockCtx.EXPECT().OptionBool("with-secret").Return(false).Once()
	mockCtx.EXPECT().Success("The application is in maintenance mode now").Once()
	s.mockConfig.EXPECT().GetString("APP_MAINTENANCE_DRIVER", "file").Return("file").Once()
	s.mockStorage.EXPECT().Put("framework/maintenance.json", "{\"render\":\"errors/503.tmpl\",\"status\":503}").Return(nil).Once()

	cmd := NewDownCommand(s.mockView, s.mockHash, s.maintenance())
	err := cmd.Handle(mockCtx)

	assert.Nil(s.T(), err)
}
func (s *DownCommandTestSuite) TestHandleSecret() {
	s.mockHash.EXPECT().Make("secretpassword").Return("hashedsecretpassword", nil)

	mockCtx := mocksconsole.NewContext(s.T())
	mockCtx.EXPECT().OptionInt("status").Return(503).Once()
	mockCtx.EXPECT().Option("reason").Return("Under maintenance").Once()
	mockCtx.EXPECT().Option("render").Return("").Once()
	mockCtx.EXPECT().Option("redirect").Return("").Once()
	mockCtx.EXPECT().Option("secret").Return("secretpassword").Once()
	mockCtx.EXPECT().OptionBool("with-secret").Return(false).Once()
	mockCtx.EXPECT().Success("The application is in maintenance mode now").Once()
	s.mockConfig.EXPECT().GetString("APP_MAINTENANCE_DRIVER", "file").Return("file").Once()
	s.mockStorage.EXPECT().Put("framework/maintenance.json", "{\"reason\":\"Under maintenance\",\"secret\":\"hashedsecretpassword\",\"status\":503}").Return(nil).Once()

	cmd := NewDownCommand(s.mockView, s.mockHash, s.maintenance())
	err := cmd.Handle(mockCtx)

	assert.Nil(s.T(), err)
}

func (s *DownCommandTestSuite) TestHandleSecretWithNilHash() {
	mockCtx := mocksconsole.NewContext(s.T())
	mockCtx.EXPECT().OptionInt("status").Return(503).Once()
	mockCtx.EXPECT().Option("render").Return("").Once()
	mockCtx.EXPECT().Option("redirect").Return("").Once()
	mockCtx.EXPECT().Option("reason").Return("Under maintenance").Once()
	mockCtx.EXPECT().Option("secret").Return("secretpassword").Once()
	mockCtx.EXPECT().OptionBool("with-secret").Return(false).Once()
	mockCtx.EXPECT().Error(frameworkerrors.HashFacadeNotSet.Error()).Once()

	cmd := NewDownCommand(s.mockView, nil, s.maintenance())
	err := cmd.Handle(mockCtx)

	assert.Nil(s.T(), err)
}

func (s *DownCommandTestSuite) TestHandleWithSecret() {
	s.mockHash.EXPECT().Make(mock.Anything).Return("randomhashedsecretpassword", nil)

	mockCtx := mocksconsole.NewContext(s.T())
	mockCtx.EXPECT().OptionInt("status").Return(503).Once()
	mockCtx.EXPECT().Option("reason").Return("Under maintenance").Once()
	mockCtx.EXPECT().Option("render").Return("").Once()
	mockCtx.EXPECT().Option("redirect").Return("").Once()
	mockCtx.EXPECT().Option("secret").Return("").Once()
	mockCtx.EXPECT().OptionBool("with-secret").Return(true).Once()
	mockCtx.EXPECT().Info(mock.MatchedBy(func(arg string) bool {
		return strings.HasPrefix(arg, "Using secret:")
	})).Once()
	mockCtx.EXPECT().Success("The application is in maintenance mode now").Once()
	s.mockConfig.EXPECT().GetString("APP_MAINTENANCE_DRIVER", "file").Return("file").Once()
	s.mockStorage.EXPECT().Put("framework/maintenance.json", "{\"reason\":\"Under maintenance\",\"secret\":\"randomhashedsecretpassword\",\"status\":503}").Return(nil).Once()

	cmd := NewDownCommand(s.mockView, s.mockHash, s.maintenance())
	err := cmd.Handle(mockCtx)

	assert.Nil(s.T(), err)
}

func (s *DownCommandTestSuite) TestHandleWithSecretWithNilHash() {
	mockCtx := mocksconsole.NewContext(s.T())
	mockCtx.EXPECT().OptionInt("status").Return(503).Once()
	mockCtx.EXPECT().Option("render").Return("").Once()
	mockCtx.EXPECT().Option("redirect").Return("").Once()
	mockCtx.EXPECT().Option("reason").Return("Under maintenance").Once()
	mockCtx.EXPECT().Option("secret").Return("").Once()
	mockCtx.EXPECT().OptionBool("with-secret").Return(true).Once()
	mockCtx.EXPECT().Error(frameworkerrors.HashFacadeNotSet.Error()).Once()

	cmd := NewDownCommand(s.mockView, nil, s.maintenance())
	err := cmd.Handle(mockCtx)

	assert.Nil(s.T(), err)
}

func (s *DownCommandTestSuite) TestHandleWithCacheDriver() {
	mockCtx := mocksconsole.NewContext(s.T())
	mockCtx.EXPECT().OptionInt("status").Return(503).Once()
	mockCtx.EXPECT().Option("reason").Return("Under maintenance").Once()
	mockCtx.EXPECT().Option("render").Return("").Once()
	mockCtx.EXPECT().Option("redirect").Return("").Once()
	mockCtx.EXPECT().Option("secret").Return("").Once()
	mockCtx.EXPECT().OptionBool("with-secret").Return(false).Once()
	mockCtx.EXPECT().Success("The application is in maintenance mode now").Once()
	s.mockConfig.EXPECT().GetString("APP_MAINTENANCE_DRIVER", "file").Return("cache").Once()
	s.mockConfig.EXPECT().GetString("APP_MAINTENANCE_STORE").Return("").Once()
	s.mockCache.EXPECT().Forever("framework:maintenance", "{\"reason\":\"Under maintenance\",\"status\":503}").Return(true).Once()

	cmd := NewDownCommand(s.mockView, s.mockHash, s.maintenance())
	err := cmd.Handle(mockCtx)

	assert.Nil(s.T(), err)
}

func (s *DownCommandTestSuite) TestHandleWithNamedCacheStore() {
	mockCacheDriver := mockscache.NewDriver(s.T())
	mockCtx := mocksconsole.NewContext(s.T())
	mockCtx.EXPECT().OptionInt("status").Return(503).Once()
	mockCtx.EXPECT().Option("reason").Return("Under maintenance").Once()
	mockCtx.EXPECT().Option("render").Return("").Once()
	mockCtx.EXPECT().Option("redirect").Return("").Once()
	mockCtx.EXPECT().Option("secret").Return("").Once()
	mockCtx.EXPECT().OptionBool("with-secret").Return(false).Once()
	mockCtx.EXPECT().Success("The application is in maintenance mode now").Once()
	s.mockConfig.EXPECT().GetString("APP_MAINTENANCE_DRIVER", "file").Return("cache").Once()
	s.mockConfig.EXPECT().GetString("APP_MAINTENANCE_STORE").Return("redis").Once()
	s.mockCache.EXPECT().Store("redis").Return(mockCacheDriver).Once()
	mockCacheDriver.EXPECT().Forever("framework:maintenance", "{\"reason\":\"Under maintenance\",\"status\":503}").Return(true).Once()

	cmd := NewDownCommand(s.mockView, s.mockHash, s.maintenance())
	err := cmd.Handle(mockCtx)

	assert.Nil(s.T(), err)
}
