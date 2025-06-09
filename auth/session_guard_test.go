package auth

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/errors"
	mocksauth "github.com/goravel/framework/mocks/auth"
	mockscache "github.com/goravel/framework/mocks/cache"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
	mockslog "github.com/goravel/framework/mocks/log"
	"github.com/goravel/framework/support/carbon"
)

type SessionGuardTestSuite struct {
	suite.Suite
	sessionGuard     *SessionGuard
	mockCache        *mockscache.Cache
	mockConfig       *mocksconfig.Config
	mockContext      http.Context
	mockDB           *mocksorm.Query
	mockLog          *mockslog.Log
	mockUserProvider *mocksauth.UserProvider
	now              *carbon.Carbon
}

func TestSessionGuardTestSuite(t *testing.T) {
	suite.Run(t, new(SessionGuardTestSuite))
}

func (s *SessionGuardTestSuite) SetupTest() {
	s.mockCache = mockscache.NewCache(s.T())
	s.mockConfig = mocksconfig.NewConfig(s.T())
	s.mockContext = Background(s.T())
	s.mockDB = mocksorm.NewQuery(s.T())
	s.mockLog = mockslog.NewLog(s.T())
	s.mockUserProvider = mocksauth.NewUserProvider(s.T())

	cacheFacade = s.mockCache
	configFacade = s.mockConfig

	sessionGuard, err := NewSessionGuard(s.mockContext, testUserGuard, s.mockUserProvider)
	s.Require().Nil(err)

	now := carbon.Now()
	carbon.SetTestNow(now)
	s.now = now
	s.sessionGuard = sessionGuard.(*SessionGuard)
}

func (s *SessionGuardTestSuite) TestLoginUsingID_InvalidKey() {
	token, err := s.sessionGuard.LoginUsingID("")
	s.Empty(token)
	s.ErrorIs(err, errors.AuthInvalidKey)
}
