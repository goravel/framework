package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/errors"
	mocksauth "github.com/goravel/framework/mocks/auth"
	mockscache "github.com/goravel/framework/mocks/cache"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
	mockshttp "github.com/goravel/framework/mocks/http"
	mockslog "github.com/goravel/framework/mocks/log"
	mockssession "github.com/goravel/framework/mocks/session"
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
	mockSession      *mockssession.Session
	now              *carbon.Carbon
}

func TestSessionGuardTestSuite(t *testing.T) {
	suite.Run(t, new(SessionGuardTestSuite))
}

func (s *SessionGuardTestSuite) TearDownSuite() {
	carbon.ClearTestNow()
}

func (s *SessionGuardTestSuite) SetupTest() {
	s.mockCache = mockscache.NewCache(s.T())
	s.mockConfig = mocksconfig.NewConfig(s.T())
	s.mockDB = mocksorm.NewQuery(s.T())
	s.mockLog = mockslog.NewLog(s.T())
	s.mockUserProvider = mocksauth.NewUserProvider(s.T())

	s.mockSession = mockssession.NewSession(s.T())

	mockRequest := mockshttp.NewContextRequest(s.T())
	mockRequest.EXPECT().Session().Return(s.mockSession)

	mockContext := mockshttp.NewContext(s.T())
	mockContext.EXPECT().Request().Return(mockRequest)

	s.mockContext = mockContext

	cacheFacade = s.mockCache
	configFacade = s.mockConfig

	sessionGuard, err := NewSessionGuard(s.mockContext, testUserGuard, s.mockUserProvider)
	s.Require().Nil(err)

	now := carbon.Now()
	carbon.SetTestNow(now)
	s.now = now
	s.sessionGuard = sessionGuard.(*SessionGuard)
}

func (s *SessionGuardTestSuite) TestNewSessionGuard() {
	sessionGuard, err := NewSessionGuard(nil, testUserGuard, s.mockUserProvider)

	s.Nil(sessionGuard)
	s.NotNil(err)
	s.ErrorIs(err, errors.InvalidHttpContext)

	mockRequest := mockshttp.NewContextRequest(s.T())
	mockRequest.EXPECT().Session().Return(nil).Once()

	mockContext := mockshttp.NewContext(s.T())
	mockContext.EXPECT().Request().Return(mockRequest).Once()

	s.mockContext = mockContext
	sessionGuard, err = NewSessionGuard(s.mockContext, testUserGuard, s.mockUserProvider)

	s.Nil(sessionGuard)
	s.NotNil(err)
	s.ErrorIs(err, errors.SessionDriverIsNotSet)

	mockRequest.EXPECT().Session().Return(s.mockSession)
	mockContext.EXPECT().Request().Return(mockRequest).Once()
	sessionGuard, err = NewSessionGuard(s.mockContext, testUserGuard, s.mockUserProvider)

	s.Nil(err)
	s.NotNil(sessionGuard)
}

func (s *SessionGuardTestSuite) TestLoginUsingID_InvalidKey() {
	token, err := s.sessionGuard.LoginUsingID("")
	s.Empty(token)
	s.ErrorIs(err, errors.AuthInvalidKey)
}

func (s *SessionGuardTestSuite) TestCheck_LoginUsingID_Logout() {
	s.mockSession.EXPECT().Get("auth_user_id", nil).Return(nil).Twice()
	s.False(s.sessionGuard.Check())
	s.True(s.sessionGuard.Guest())

	s.mockSession.EXPECT().Put("auth_user_id", "1").Return(nil).Once()
	token, err := s.sessionGuard.LoginUsingID(1)
	s.Nil(err)
	s.Empty(token)

	s.mockSession.EXPECT().Get("auth_user_id", nil).Return("1").Twice()
	s.True(s.sessionGuard.Check())
	s.False(s.sessionGuard.Guest())

	s.mockSession.EXPECT().Forget("auth_user_id").Return(nil).Once()
	s.NoError(s.sessionGuard.Logout())

	s.mockSession.EXPECT().Get("auth_user_id", nil).Return(nil).Once()
	s.True(s.sessionGuard.Guest())
}

func (s *SessionGuardTestSuite) Test_Login() {
	s.mockSession.EXPECT().Get("auth_user_id", nil).Return(nil).Twice()
	s.False(s.sessionGuard.Check())
	s.True(s.sessionGuard.Guest())

	var user User
	user.ID = 2
	user.Name = "Goravel"

	s.mockUserProvider.EXPECT().GetID(&user).Return("2", nil).Once()
	s.mockSession.EXPECT().Put("auth_user_id", "2").Return(nil).Once()
	token, err := s.sessionGuard.Login(&user)
	s.Nil(err)
	s.Empty(token)

	s.mockSession.EXPECT().Get("auth_user_id", nil).Return("2").Twice()
	s.True(s.sessionGuard.Check())
	s.False(s.sessionGuard.Guest())

	s.mockSession.EXPECT().Forget("auth_user_id").Return(nil).Once()
	s.NoError(s.sessionGuard.Logout())

	s.mockSession.EXPECT().Get("auth_user_id", nil).Return(nil).Once()
	s.True(s.sessionGuard.Guest())
}

func (s *SessionGuardTestSuite) Test_LoginFailed() {
	s.mockSession.EXPECT().Get("auth_user_id", nil).Return(nil).Twice()
	s.False(s.sessionGuard.Check())
	s.True(s.sessionGuard.Guest())

	var user User
	user.ID = 2
	user.Name = "Goravel"

	s.mockUserProvider.EXPECT().GetID(&user).Return("", assert.AnError).Once()
	token, err := s.sessionGuard.Login(&user)
	s.NotNil(err)
	s.Empty(token)

	s.mockSession.EXPECT().Get("auth_user_id", nil).Return(nil).Twice()
	s.False(s.sessionGuard.Check())
	s.True(s.sessionGuard.Guest())

	s.mockSession.EXPECT().Forget("auth_user_id").Return(nil).Once()
	s.NoError(s.sessionGuard.Logout())

	s.mockSession.EXPECT().Get("auth_user_id", nil).Return(nil).Once()
	s.True(s.sessionGuard.Guest())
}

func (s *SessionGuardTestSuite) Test_User() {
	var user User

	s.mockUserProvider.EXPECT().RetriveByID(&user, "1").RunAndReturn(func(user any, id any) error {
		user.(*User).ID = 1
		user.(*User).Name = "Goravel"
		return nil
	}).Once()

	s.mockSession.EXPECT().Get("auth_user_id", nil).Return("1").Once()

	err := s.sessionGuard.User(&user)

	var id uint = 1

	s.Nil(err)
	s.Equal(id, user.ID)
	s.Equal("Goravel", user.Name)
}

func (s *SessionGuardTestSuite) Test_Parse() {
	token, err := s.sessionGuard.Parse("")
	s.Empty(token)
	s.NotNil(err)
	s.EqualError(err, "The method was not supported for the driver session")
}

func (s *SessionGuardTestSuite) Test_Refresh() {
	token, err := s.sessionGuard.Refresh()
	s.Empty(token)
	s.NotNil(err)
	s.EqualError(err, "The method was not supported for the driver session")
}
