package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	contractsauth "github.com/goravel/framework/contracts/auth"
	"github.com/goravel/framework/contracts/http"
	mocksauth "github.com/goravel/framework/mocks/auth"
	mockscache "github.com/goravel/framework/mocks/cache"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
	mockslog "github.com/goravel/framework/mocks/log"
)

type AuthTestSuite struct {
	suite.Suite
	auth        *Auth
	mockCache   *mockscache.Cache
	mockConfig  *mocksconfig.Config
	mockContext http.Context
	mockOrm     *mocksorm.Orm
	mockDB      *mocksorm.Query
	mockLog     *mockslog.Log
}

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}

func (s *AuthTestSuite) SetupTest() {
	s.mockCache = mockscache.NewCache(s.T())
	s.mockConfig = mocksconfig.NewConfig(s.T())
	s.mockContext = Background()
	s.mockOrm = mocksorm.NewOrm(s.T())
	s.mockDB = mocksorm.NewQuery(s.T())
	s.mockLog = mockslog.NewLog(s.T())

	cacheFacade = s.mockCache
	configFacade = s.mockConfig
	ormFacade = s.mockOrm

	s.mockConfig.EXPECT().GetString("auth.defaults.guard").Return("user").Once()

	s.mockConfig.EXPECT().GetString("auth.guards.user.driver").Return("jwt").Once()
	s.mockConfig.EXPECT().GetString("auth.guards.user.provider").Return("user").Once()
	s.mockConfig.EXPECT().GetInt("auth.guards.user.ttl").Return(2).Once()
	s.mockConfig.EXPECT().GetString("auth.providers.user.driver").Return("orm").Once()

	s.mockConfig.EXPECT().GetString("auth.guards.user.secret").Return("a").Once()
	s.mockConfig.EXPECT().GetInt("auth.guards.user.refresh_ttl").Return(2).Once()

	auth, err := NewAuth(s.mockContext, s.mockConfig, s.mockLog)
	s.Require().Nil(err)
	s.auth = auth
}

func (s *AuthTestSuite) TestCustomGuardAndProvider() {
	mockProvider := mocksauth.NewUserProvider(s.T())
	s.auth.Extend("session", func(ctx http.Context, name string, userProvider contractsauth.UserProvider) (contractsauth.GuardDriver, error) {
		mockGuard := mocksauth.NewGuardDriver(s.T())
		mockGuard.EXPECT().ID().Return("session-id-xxxx", nil)
		return mockGuard, nil
	})
	s.auth.Provider("mock", func(ctx http.Context) (contractsauth.UserProvider, error) {
		return mockProvider, nil
	})

	s.mockConfig.EXPECT().GetString("auth.guards.admin.driver").Return("session").Once()
	s.mockConfig.EXPECT().GetString("auth.guards.admin.provider").Return("admin").Once()
	s.mockConfig.EXPECT().GetString("auth.providers.admin.driver").Return("mock").Once()

	guard := s.auth.Guard("admin")

	id, err := guard.ID()
	s.Nil(err)

	s.Equal("session-id-xxxx", id)
}

func (s *AuthTestSuite) TestUserProviderReturnsError() {
	s.auth.Extend("session", func(ctx http.Context, name string, userProvider contractsauth.UserProvider) (contractsauth.GuardDriver, error) {
		mockGuard := mocksauth.NewGuardDriver(s.T())
		mockGuard.EXPECT().ID().Return("session-id-xxxx", nil)
		return mockGuard, nil
	})
	s.auth.Provider("mock", func(ctx http.Context) (contractsauth.UserProvider, error) {
		return nil, assert.AnError
	})

	s.mockConfig.EXPECT().GetString("auth.guards.admin.driver").Return("session").Once()
	s.mockConfig.EXPECT().GetString("auth.guards.admin.provider").Return("admin").Once()
	s.mockConfig.EXPECT().GetString("auth.providers.admin.driver").Return("mock").Once()

	s.Panics(func() {
		guard := s.auth.Guard("admin")
		s.Nil(guard)
	})
}

func (s *AuthTestSuite) TestGuardDriverReturnsError() {
	mockProvider := mocksauth.NewUserProvider(s.T())
	s.auth.Extend("session", func(ctx http.Context, name string, userProvider contractsauth.UserProvider) (contractsauth.GuardDriver, error) {
		return nil, assert.AnError
	})
	s.auth.Provider("mock", func(ctx http.Context) (contractsauth.UserProvider, error) {
		return mockProvider, nil
	})

	s.mockConfig.EXPECT().GetString("auth.guards.admin.driver").Return("session").Once()
	s.mockConfig.EXPECT().GetString("auth.guards.admin.provider").Return("admin").Once()
	s.mockConfig.EXPECT().GetString("auth.providers.admin.driver").Return("mock").Once()

	s.Panics(func() {
		guard := s.auth.Guard("admin")
		s.Nil(guard)
	})
}

func (s *AuthTestSuite) TestGuardDriverNotFound() {
	s.mockConfig.EXPECT().GetString("auth.guards.admin.driver").Return("unknown").Once()

	s.Panics(func() {
		guard := s.auth.Guard("admin")
		s.Nil(guard)
	})
}

func (s *AuthTestSuite) TestUserProviderDriverNotFound() {
	s.mockConfig.EXPECT().GetString("auth.guards.admin.driver").Return("jwt").Once()
	s.mockConfig.EXPECT().GetString("auth.guards.admin.provider").Return("admin").Once()
	s.mockConfig.EXPECT().GetString("auth.providers.admin.driver").Return("unknown").Once()

	s.Panics(func() {
		guard := s.auth.Guard("admin")
		s.Nil(guard)
	})
}
