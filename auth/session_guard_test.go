package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	"github.com/goravel/framework/session"
	"github.com/goravel/framework/support/carbon"
)

type SessionGuardTestSuite struct {
	suite.Suite
	sessionGuard     *SessionGuard
	mockCache        *mockscache.Cache
	mockConfig       *mocksconfig.Config
	mockContext      http.Context
	mockResponse     *mockshttp.ContextResponse
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

	s.mockResponse = mockshttp.NewContextResponse(s.T())
	mockContext := mockshttp.NewContext(s.T())
	mockContext.EXPECT().Request().Return(mockRequest)
	mockContext.EXPECT().Response().Return(s.mockResponse).Maybe()

	s.mockContext = mockContext

	cacheFacade = s.mockCache
	configFacade = s.mockConfig
	session.ConfigFacade = s.mockConfig

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

	s.mockSession.EXPECT().Regenerate(true).Return(nil).Once()
	s.expectReissueCookie()
	s.mockSession.EXPECT().Put("auth_user_id", "1").Return(nil).Once()
	token, err := s.sessionGuard.LoginUsingID(1)
	s.Nil(err)
	s.Empty(token)

	s.mockSession.EXPECT().Get("auth_user_id", nil).Return("1").Twice()
	s.True(s.sessionGuard.Check())
	s.False(s.sessionGuard.Guest())

	s.mockSession.EXPECT().Invalidate().Return(nil).Once()
	s.expectReissueCookie()
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
	s.mockSession.EXPECT().Regenerate(true).Return(nil).Once()
	s.expectReissueCookie()
	s.mockSession.EXPECT().Put("auth_user_id", "2").Return(nil).Once()
	token, err := s.sessionGuard.Login(&user)
	s.Nil(err)
	s.Empty(token)

	s.mockSession.EXPECT().Get("auth_user_id", nil).Return("2").Twice()
	s.True(s.sessionGuard.Check())
	s.False(s.sessionGuard.Guest())

	s.mockSession.EXPECT().Invalidate().Return(nil).Once()
	s.expectReissueCookie()
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

	s.mockSession.EXPECT().Invalidate().Return(nil).Once()
	s.expectReissueCookie()
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

func (s *SessionGuardTestSuite) Test_InvalidKey() {
	var user User
	s.mockSession.EXPECT().Get("auth_user_id", nil).Return(user).Once()

	err := s.sessionGuard.User(&user)

	s.NotNil(err)
	s.ErrorIs(err, errors.AuthInvalidKey)
}

func (s *SessionGuardTestSuite) Test_LoginUsingID_RegenerateError() {
	s.mockSession.EXPECT().Regenerate(true).Return(assert.AnError).Once()

	token, err := s.sessionGuard.LoginUsingID(1)
	s.Empty(token)
	s.ErrorIs(err, assert.AnError)
}

func (s *SessionGuardTestSuite) Test_Logout_InvalidateError() {
	s.mockSession.EXPECT().Invalidate().Return(assert.AnError).Once()

	s.ErrorIs(s.sessionGuard.Logout(), assert.AnError)
}

func (s *SessionGuardTestSuite) Test_LoginUsingID_ReissuesCookie() {
	s.mockSession.EXPECT().Regenerate(true).Return(nil).Once()
	s.mockSession.EXPECT().GetName().Return("goravel_session").Once()
	s.mockSession.EXPECT().GetID().Return("new-session-id").Once()
	s.mockSession.EXPECT().Put("auth_user_id", "1").Return(nil).Once()

	s.mockConfig.EXPECT().GetInt("session.lifetime", 120).Return(120).Once()
	s.mockConfig.EXPECT().GetString("session.path").Return("/").Once()
	s.mockConfig.EXPECT().GetString("session.domain").Return("").Once()
	s.mockConfig.EXPECT().GetBool("session.secure").Return(false).Once()
	s.mockConfig.EXPECT().GetBool("session.http_only").Return(true).Once()
	s.mockConfig.EXPECT().GetString("session.same_site").Return("").Once()

	s.mockResponse.EXPECT().Cookie(mock.MatchedBy(func(c http.Cookie) bool {
		return c.Name == "goravel_session" && c.Value == "new-session-id"
	})).Return(s.mockResponse).Once()

	_, err := s.sessionGuard.LoginUsingID(1)
	s.Nil(err)
}

func (s *SessionGuardTestSuite) expectReissueCookie() {
	s.mockSession.EXPECT().GetName().Return("goravel_session").Once()
	s.mockSession.EXPECT().GetID().Return("session-id").Once()
	s.mockConfig.EXPECT().GetInt("session.lifetime", 120).Return(120).Once()
	s.mockConfig.EXPECT().GetString("session.path").Return("/").Once()
	s.mockConfig.EXPECT().GetString("session.domain").Return("").Once()
	s.mockConfig.EXPECT().GetBool("session.secure").Return(false).Once()
	s.mockConfig.EXPECT().GetBool("session.http_only").Return(true).Once()
	s.mockConfig.EXPECT().GetString("session.same_site").Return("").Once()
	s.mockResponse.EXPECT().Cookie(mock.Anything).Return(s.mockResponse).Once()
}
