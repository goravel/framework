package auth

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	contractsauth "github.com/goravel/framework/contracts/auth"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/errors"
	mocksauth "github.com/goravel/framework/mocks/auth"
	mockscache "github.com/goravel/framework/mocks/cache"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
	mockslog "github.com/goravel/framework/mocks/log"
	"github.com/goravel/framework/support/carbon"
)

type JwtGuardTestSuite struct {
	suite.Suite
	jwtGuard         *JwtGuard
	mockCache        *mockscache.Cache
	mockConfig       *mocksconfig.Config
	mockContext      http.Context
	mockDB           *mocksorm.Query
	mockLog          *mockslog.Log
	mockUserProvider *mocksauth.UserProvider
	now              *carbon.Carbon
}

func TestJwtGuardTestSuite(t *testing.T) {
	suite.Run(t, new(JwtGuardTestSuite))
}

func (s *JwtGuardTestSuite) SetupTest() {
	s.mockCache = mockscache.NewCache(s.T())
	s.mockConfig = mocksconfig.NewConfig(s.T())
	s.mockContext = Background()
	s.mockDB = mocksorm.NewQuery(s.T())
	s.mockLog = mockslog.NewLog(s.T())
	s.mockUserProvider = mocksauth.NewUserProvider(s.T())

	cacheFacade = s.mockCache
	configFacade = s.mockConfig

	s.mockConfig.EXPECT().GetString("auth.guards.user.secret").Return("a").Once()
	s.mockConfig.EXPECT().GetInt("auth.guards.user.refresh_ttl").Return(2).Once()
	s.mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(2).Once()

	jwtGuard, err := NewJwtGuard(s.mockContext, testUserGuard, s.mockUserProvider)
	s.Require().Nil(err)

	now := carbon.Now()
	carbon.SetTestNow(now)
	s.now = now
	s.jwtGuard = jwtGuard.(*JwtGuard)
}

func (s *JwtGuardTestSuite) TestLoginUsingID_InvalidKey() {
	token, err := s.jwtGuard.LoginUsingID("")
	s.Empty(token)
	s.ErrorIs(err, errors.AuthInvalidKey)
}

func (s *JwtGuardTestSuite) TestCheck_LoginUsingID_Logout() {
	s.mockCache.EXPECT().Put(mock.MatchedBy(func(key string) bool {
		return strings.HasPrefix(key, "jwt:disabled:")
	}), true, 2*time.Minute).Return(nil).Once()

	s.False(s.jwtGuard.Check())
	s.True(s.jwtGuard.Guest())

	token, err := s.jwtGuard.LoginUsingID(1)
	s.Nil(err)
	s.NotEmpty(token)

	s.True(s.jwtGuard.Check())
	s.False(s.jwtGuard.Guest())
	s.NoError(s.jwtGuard.Logout())
	s.True(s.jwtGuard.Guest())
}

func (s *JwtGuardTestSuite) TestLogin() {
	var user User
	user.ID = 1
	user.Name = "Goravel"

	s.mockUserProvider.EXPECT().GetID(&user).Return(1, nil).Once()

	token, err := s.jwtGuard.Login(&user)
	s.Nil(err)
	s.NotEmpty(token)
}

func (s *JwtGuardTestSuite) TestLogin_Failed() {
	var user User

	s.mockUserProvider.EXPECT().GetID(&user).Return(nil, assert.AnError).Once()

	token, err := s.jwtGuard.Login(&user)
	s.Empty(token)
	s.EqualError(err, assert.AnError.Error())
}

func (s *JwtGuardTestSuite) TestParse_TokenDisabled() {
	token := "1"
	s.mockCache.EXPECT().GetBool("jwt:disabled:"+token, false).Return(true).Once()

	payload, err := s.jwtGuard.Parse(token)
	s.Nil(payload)
	s.EqualError(err, errors.AuthTokenDisabled.Error())
}

func (s *JwtGuardTestSuite) TestParse_TokenInvalid() {
	token := "1"
	s.mockCache.EXPECT().GetBool("jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.jwtGuard.Parse(token)
	s.Nil(payload)
	s.ErrorIs(err, errors.AuthInvalidToken)
}

func (s *JwtGuardTestSuite) TestParse_TokenExpired() {
	issuedAt := s.now.StdTime()
	expireAt := s.now.Copy().AddMinutes(2).StdTime()

	token, err := s.jwtGuard.LoginUsingID(1)
	s.Nil(err)
	s.NotEmpty(token)

	carbon.SetTestNow(s.now.Copy().AddMinutes(2))

	guardInfo, err := s.jwtGuard.GetJwtToken()
	s.Nil(err)

	s.mockCache.EXPECT().GetBool("jwt:disabled:"+guardInfo.Token, false).Return(false).Once()

	payload, err := s.jwtGuard.Parse(guardInfo.Token)
	s.Equal(&contractsauth.Payload{
		Guard:    testUserGuard,
		Key:      "1",
		ExpireAt: jwt.NewNumericDate(expireAt).Local(),
		IssuedAt: jwt.NewNumericDate(issuedAt).Local(),
	}, payload)
	s.ErrorIs(err, errors.AuthTokenExpired)

	carbon.ClearTestNow()
}

func (s *JwtGuardTestSuite) TestParse_Success() {
	token, err := s.jwtGuard.LoginUsingID(1)
	s.Nil(err)
	s.NotEmpty(token)

	s.mockCache.EXPECT().GetBool("jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.jwtGuard.Parse(token)
	s.NoError(err)

	s.Equal(&contractsauth.Payload{
		Guard:    testUserGuard,
		Key:      "1",
		ExpireAt: jwt.NewNumericDate(s.now.Copy().AddMinutes(2).StdTime()).Local(),
		IssuedAt: jwt.NewNumericDate(s.now.Copy().StdTime()).Local(),
	}, payload)
}

func (s *JwtGuardTestSuite) TestUser_NoParse() {
	var user User
	err := s.jwtGuard.User(user)

	s.EqualError(err, errors.AuthParseTokenFirst.Error())
}

func (s *JwtGuardTestSuite) TestID_NoParse() {
	id, err := s.jwtGuard.ID()

	s.ErrorIs(err, errors.AuthParseTokenFirst)
	s.Empty(id)
}

func (s *JwtGuardTestSuite) TestID_Success() {
	token, err := s.jwtGuard.LoginUsingID(1)
	s.Nil(err)
	s.NotEmpty(token)

	s.mockCache.EXPECT().GetBool("jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.jwtGuard.Parse(token)
	s.Nil(err)
	s.NotNil(payload)

	id, err := s.jwtGuard.ID()
	s.Nil(err)
	s.Equal("1", id)
}

func (s *JwtGuardTestSuite) TestID_TokenExpired() {
	token, err := s.jwtGuard.LoginUsingID(1)
	s.NotEmpty(token)
	s.Nil(err)

	// Set the token as expired
	carbon.SetTestNow(s.now.AddMinutes(3))

	s.mockCache.EXPECT().GetBool("jwt:disabled:"+token, false).Return(false).Once()

	// Parse the token
	_, err = s.jwtGuard.Parse(token)
	s.ErrorIs(err, errors.AuthTokenExpired)

	// Now, call the ID method and expect it to return an empty value
	id, err := s.jwtGuard.ID()
	s.Empty(id)
	s.ErrorIs(err, errors.AuthTokenExpired)

	carbon.ClearTestNow()
}

func (s *JwtGuardTestSuite) TestID_TokenInvalid() {
	token := "invalidToken"
	s.mockCache.EXPECT().GetBool("jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.jwtGuard.Parse(token)
	s.Nil(payload)
	s.ErrorIs(err, errors.AuthInvalidToken)

	id, err := s.jwtGuard.ID()
	s.Empty(id)
	s.ErrorIs(err, errors.AuthParseTokenFirst)
}

func (s *JwtGuardTestSuite) TestUser_Failed() {
	token, err := s.jwtGuard.LoginUsingID(1)
	s.Nil(err)
	s.NotEmpty(token)

	s.mockCache.EXPECT().GetBool("jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.jwtGuard.Parse(token)
	s.NotNil(payload)
	s.Nil(err)

	var user User

	s.mockUserProvider.EXPECT().RetriveByID(&user, "1").Return(assert.AnError).Once()

	err = s.jwtGuard.User(&user)
	s.EqualError(err, assert.AnError.Error())
}

func (s *JwtGuardTestSuite) TestUser_Expired_Refresh() {
	token, err := s.jwtGuard.LoginUsingID(1)
	s.Nil(err)
	s.NotEmpty(token)

	s.mockCache.EXPECT().GetBool("jwt:disabled:"+token, false).Return(false).Once()

	carbon.SetTestNow(s.now.Copy().AddMinutes(2))

	payload, err := s.jwtGuard.Parse(token)
	s.NotNil(payload)
	s.ErrorIs(err, errors.AuthTokenExpired)

	var user User
	err = s.jwtGuard.User(&user)
	s.EqualError(err, errors.AuthTokenExpired.Error())

	token, err = s.jwtGuard.Refresh()
	s.NotEmpty(token)
	s.Nil(err)

	s.mockUserProvider.EXPECT().RetriveByID(&user, "1").Return(nil).Once()

	err = s.jwtGuard.User(&user)
	s.Nil(err)
}

func (s *JwtGuardTestSuite) TestUser_RefreshExpired() {
	token, err := s.jwtGuard.LoginUsingID(1)
	s.Nil(err)
	s.NotEmpty(token)

	s.mockCache.EXPECT().GetBool("jwt:disabled:"+token, false).Return(false).Once()

	carbon.SetTestNow(s.now.Copy().AddMinutes(2))

	payload, err := s.jwtGuard.Parse(token)
	s.NotNil(payload)
	s.ErrorIs(err, errors.AuthTokenExpired)

	var user User
	err = s.jwtGuard.User(&user)
	s.EqualError(err, errors.AuthTokenExpired.Error())

	carbon.SetTestNow(s.now.Copy().AddMinutes(5))

	token, err = s.jwtGuard.Refresh()
	s.Empty(token)
	s.EqualError(err, errors.AuthRefreshTimeExceeded.Error())

	carbon.ClearTestNow()
}

func (s *JwtGuardTestSuite) TestUser_Success() {
	token, err := s.jwtGuard.LoginUsingID(1)
	s.Nil(err)
	s.NotEmpty(token)

	s.mockCache.EXPECT().GetBool("jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.jwtGuard.Parse(token)
	s.NotNil(payload)
	s.Nil(err)

	var user User
	s.mockUserProvider.EXPECT().RetriveByID(&user, "1").RunAndReturn(func(user interface{}, id interface{}) error {
		user.(*User).ID = 1
		return nil
	}).Once()

	err = s.jwtGuard.User(&user)
	s.Nil(err)
	s.Equal(uint(1), user.ID)
}

func (s *JwtGuardTestSuite) TestUser_Success_MultipleParse() {
	testAdminGuard := "admin"

	s.mockConfig.EXPECT().Get("auth.guards.admin.ttl").Return(2)
	s.mockConfig.EXPECT().GetString("auth.guards.admin.secret").Return("a").Once()
	s.mockConfig.EXPECT().GetInt("auth.guards.admin.refresh_ttl").Return(0).Once()
	s.mockConfig.EXPECT().GetInt("jwt.refresh_ttl").Return(2).Once()

	adminJwtGuard, err := NewJwtGuard(s.mockContext, testAdminGuard, s.mockUserProvider)
	s.Require().Nil(err)

	userToken, err := s.jwtGuard.LoginUsingID(1)
	s.NoError(err)
	s.NotEmpty(userToken)

	adminToken, err := adminJwtGuard.LoginUsingID(2)
	s.NoError(err)
	s.NotEmpty(adminToken)

	s.mockCache.EXPECT().GetBool("jwt:disabled:"+userToken, false).Return(false).Once()

	payload, err := s.jwtGuard.Parse(userToken)
	s.NoError(err)
	s.NotNil(payload)
	s.Equal(testUserGuard, payload.Guard)
	s.Equal("1", payload.Key)

	s.mockCache.EXPECT().GetBool("jwt:disabled:"+adminToken, false).Return(false).Once()

	payload, err = adminJwtGuard.Parse(adminToken)
	s.NoError(err)
	s.NotNil(payload)
	s.Equal(testAdminGuard, payload.Guard)
	s.Equal("2", payload.Key)

	var user1 User
	s.mockUserProvider.EXPECT().RetriveByID(&user1, "1").RunAndReturn(func(user interface{}, id interface{}) error {
		user.(*User).ID = 1
		return nil
	}).Once()

	err = s.jwtGuard.User(&user1)
	s.NoError(err)
	s.Equal(uint(1), user1.ID)

	var user2 User
	s.mockUserProvider.EXPECT().RetriveByID(&user2, "2").RunAndReturn(func(user interface{}, id interface{}) error {
		user.(*User).ID = 2
		return nil
	}).Once()

	err = adminJwtGuard.User(&user2)
	s.NoError(err)
	s.Equal(uint(2), user2.ID)
}

func (s *JwtGuardTestSuite) TestRefresh_NotParse() {
	token, err := s.jwtGuard.Refresh()
	s.Empty(token)
	s.EqualError(err, errors.AuthParseTokenFirst.Error())
}

func (s *JwtGuardTestSuite) TestLogout_NotParse() {
	s.EqualError(s.jwtGuard.Logout(), errors.AuthParseTokenFirst.Error())
	s.mockContext.WithValue(ctxJwtKey, Guards{"user": nil})
	s.EqualError(s.jwtGuard.Logout(), errors.AuthParseTokenFirst.Error())
	s.mockContext.WithValue(ctxJwtKey, Guards{"user": &JwtToken{}})
	s.EqualError(s.jwtGuard.Logout(), errors.AuthParseTokenFirst.Error())
}

func (s *JwtGuardTestSuite) TestLogout_SetDisabledCacheError() {
	token, err := s.jwtGuard.LoginUsingID(1)
	s.NoError(err)
	s.NotEmpty(token)

	s.mockCache.EXPECT().GetBool("jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.jwtGuard.Parse(token)
	s.NotNil(payload)
	s.Nil(err)

	s.mockCache.EXPECT().Put(mock.MatchedBy(func(key string) bool {
		return strings.HasPrefix(key, "jwt:disabled:")
	}), true, 2*time.Minute).Return(assert.AnError).Once()

	s.EqualError(s.jwtGuard.Logout(), assert.AnError.Error())
}

func (s *JwtGuardTestSuite) TestMakeAuthContext() {
	testAdminGuard := "admin"

	s.mockConfig.EXPECT().Get("auth.guards.admin.ttl").Return(2)
	s.mockConfig.EXPECT().GetString("auth.guards.admin.secret").Return("").Once()
	s.mockConfig.EXPECT().GetString("jwt.secret").Return("a").Once()
	s.mockConfig.EXPECT().GetInt("auth.guards.admin.refresh_ttl").Return(2).Once()

	adminJwtGuardInterface, err := NewJwtGuard(s.mockContext, testAdminGuard, s.mockUserProvider)
	s.Require().Nil(err)
	adminJwtGuard := adminJwtGuardInterface.(*JwtGuard)

	s.jwtGuard.makeAuthContext(nil, "1")
	guards, ok := s.jwtGuard.ctx.Value(ctxJwtKey).(Guards)
	s.True(ok)
	s.Equal(&JwtToken{nil, "1"}, guards[testUserGuard])

	adminJwtGuard.makeAuthContext(nil, "2")
	guards, ok = adminJwtGuard.ctx.Value(ctxJwtKey).(Guards)
	s.True(ok)
	s.Equal(&JwtToken{nil, "1"}, guards[testUserGuard])
	s.Equal(&JwtToken{nil, "2"}, guards[testAdminGuard])
}

func (s *JwtGuardTestSuite) TestRefressTtl() {
	testAdminGuard := "admin"

	s.mockConfig.EXPECT().Get("auth.guards.admin.ttl").Return(2)
	s.mockConfig.EXPECT().GetString("auth.guards.admin.secret").Return("").Once()
	s.mockConfig.EXPECT().GetString("jwt.secret").Return("a").Once()
	s.mockConfig.EXPECT().GetInt("auth.guards.admin.refresh_ttl").Return(0).Once()
	s.mockConfig.EXPECT().GetInt("jwt.refresh_ttl").Return(0).Once()

	_, err := NewJwtGuard(s.mockContext, testAdminGuard, s.mockUserProvider)
	s.Require().Nil(err)
}

func (s *JwtGuardTestSuite) TestEmptySecret() {
	testAdminGuard := "admin"

	s.mockConfig.EXPECT().GetString("auth.guards.admin.secret").Return("").Once()
	s.mockConfig.EXPECT().GetString("jwt.secret").Return("").Once()

	_, err := NewJwtGuard(s.mockContext, testAdminGuard, s.mockUserProvider)
	s.Assert().ErrorIs(errors.AuthEmptySecret, err)
}

func (s *JwtGuardTestSuite) TestCacheFacadeNotSet() {
	testAdminGuard := "admin"

	cacheFacade = nil

	_, err := NewJwtGuard(s.mockContext, testAdminGuard, s.mockUserProvider)
	s.Assert().ErrorIs(errors.CacheFacadeNotSet, err)
}

var testUserGuard = "user"

type User struct {
	ID        uint `gorm:"primaryKey" json:"id"`
	Name      string
	CreatedAt *carbon.DateTime `gorm:"autoCreateTime;column:created_at" json:"created_at"`
	UpdatedAt *carbon.DateTime `gorm:"autoUpdateTime;column:updated_at" json:"updated_at"`
}

type Context struct {
	ctx      context.Context
	request  http.ContextRequest
	response http.ContextResponse
	values   map[any]any
	mu       sync.RWMutex
}

func (r *Context) Deadline() (deadline time.Time, ok bool) {
	return r.ctx.Deadline()
}

func (r *Context) Done() <-chan struct{} {
	return r.ctx.Done()
}

func (r *Context) Err() error {
	return r.ctx.Err()
}

func (r *Context) Value(key interface{}) any {
	if k, ok := key.(string); ok {
		r.mu.RLock()
		v, ok := r.values[k]
		r.mu.RUnlock()

		if ok {
			return v
		}
	}

	return r.ctx.Value(key)
}

func (r *Context) Context() context.Context {
	return r.ctx
}

func (r *Context) WithContext(newCtx context.Context) {
	r.ctx = newCtx
}

func (r *Context) WithValue(key any, value any) {
	r.mu.Lock()
	r.values[key] = value
	r.mu.Unlock()
}

func (r *Context) Request() http.ContextRequest {
	return r.request
}

func (r *Context) Response() http.ContextResponse {
	return r.response
}

func Background() http.Context {
	return &Context{
		ctx:      context.Background(),
		request:  nil,
		response: nil,
		values:   make(map[any]any),
	}
}

func TestGetTtl(t *testing.T) {
	var mockConfig *mocksconfig.Config

	tests := []struct {
		name     string
		setup    func()
		expected int
	}{
		{
			name: "GuardTtlIsNil",
			setup: func() {
				mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(nil).Once()
				mockConfig.EXPECT().GetInt("jwt.ttl").Return(2).Once()
			},
			expected: 2,
		},
		{
			name: "GuardTtlIsNotNil",
			setup: func() {
				mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(1).Once()
			},
			expected: 1,
		},
		{
			name: "GuardTtlIsZero",
			setup: func() {
				mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(0).Once()
			},
			expected: 60 * 24 * 365 * 100,
		},
		{
			name: "JwtTtlIsZero",
			setup: func() {
				mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(nil).Once()
				mockConfig.EXPECT().GetInt("jwt.ttl").Return(0).Once()
			},
			expected: 60 * 24 * 365 * 100,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockConfig = mocksconfig.NewConfig(t)

			test.setup()

			ttl := getTtl(mockConfig, testUserGuard)
			assert.Equal(t, test.expected, ttl)
		})
	}
}
