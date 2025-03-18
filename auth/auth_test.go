package auth

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm/clause"

	contractsauth "github.com/goravel/framework/contracts/auth"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/errors"
	mocksauth "github.com/goravel/framework/mocks/auth"
	mockscache "github.com/goravel/framework/mocks/cache"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
	"github.com/goravel/framework/support/carbon"
)

var testUserGuard = "user"

type User struct {
	ID        uint `gorm:"primaryKey" json:"id"`
	Name      string
	CreatedAt carbon.DateTime `gorm:"autoCreateTime;column:created_at" json:"created_at"`
	UpdatedAt carbon.DateTime `gorm:"autoUpdateTime;column:updated_at" json:"updated_at"`
}

type Context struct {
	ctx      context.Context
	request  http.ContextRequest
	response http.ContextResponse
	values   map[any]any
	mu       sync.RWMutex
}

func (mc *Context) Deadline() (deadline time.Time, ok bool) {
	return mc.ctx.Deadline()
}

func (mc *Context) Done() <-chan struct{} {
	return mc.ctx.Done()
}

func (mc *Context) Err() error {
	return mc.ctx.Err()
}

func (mc *Context) Value(key interface{}) any {
	if k, ok := key.(string); ok {
		mc.mu.RLock()
		v, ok := mc.values[k]
		mc.mu.RUnlock()

		if ok {
			return v
		}
	}

	return mc.ctx.Value(key)
}

func (mc *Context) Context() context.Context {
	return mc.ctx
}

func (mc *Context) WithContext(newCtx context.Context) {
	mc.ctx = newCtx
}

func (mc *Context) WithValue(key any, value any) {
	mc.mu.Lock()
	mc.values[key] = value
	mc.mu.Unlock()
}

func (mc *Context) Request() http.ContextRequest {
	return mc.request
}

func (mc *Context) Response() http.ContextResponse {
	return mc.response
}

func Background() http.Context {
	return &Context{
		ctx:      context.Background(),
		request:  nil,
		response: nil,
		values:   make(map[any]any),
	}
}

type AuthTestSuite struct {
	suite.Suite
	auth        *Auth
	mockCache   *mockscache.Cache
	mockConfig  *mocksconfig.Config
	mockContext http.Context
	mockOrm     *mocksorm.Orm
	mockDB      *mocksorm.Query
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
	s.mockConfig.EXPECT().GetString("auth.guards.user.driver").Return("jwt").Once()
	s.mockConfig.EXPECT().GetString("auth.guards.user.provider").Return("user").Once()
	s.mockConfig.EXPECT().GetString("auth.providers.user.driver").Return("orm").Once()
	if auth, err := NewAuth(testUserGuard, s.mockCache, s.mockConfig, s.mockContext, s.mockOrm); err == nil {
		s.auth = auth
	}
}

func (s *AuthTestSuite) TestLoginUsingID_EmptySecret() {
	s.mockConfig.EXPECT().GetString("jwt.secret").Return("").Once()

	guard, err := s.auth.GetGuard("user")
	s.Nil(err)
	err = guard.LoginUsingID(1)
	s.ErrorIs(err, errors.AuthEmptySecret)
}

func (s *AuthTestSuite) TestLoginUsingID_InvalidKey() {
	s.mockConfig.EXPECT().GetString("jwt.secret").Return("Goravel").Once()
	s.mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(2).Once()

	guard, err := s.auth.GetGuard("user")
	s.Nil(err)
	err = guard.LoginUsingID("")
	s.ErrorIs(err, errors.AuthInvalidKey)
}

func (s *AuthTestSuite) TestLoginUsingID() {
	s.mockConfig.EXPECT().GetString("jwt.secret").Return("Goravel").Twice()

	// jwt.ttl > 0
	s.mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(2).Once()

	guard, err := s.auth.GetGuard("user")
	s.Nil(err)
	err = guard.LoginUsingID(0)
	s.Nil(err)

	// jwt.ttl == 0
	s.mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(0).Once()

	guard, err = s.auth.GetGuard("user")
	s.Nil(err)
	err = guard.LoginUsingID(1)
	s.Nil(err)
}

func (s *AuthTestSuite) TestLogin_Model() {

	s.mockConfig.EXPECT().GetString("jwt.secret").Return("Goravel").Once()
	s.mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(2).Once()

	var user User
	user.ID = 1
	user.Name = "Goravel"
	err := s.auth.Login(&user)
	s.Nil(err)
}

func (s *AuthTestSuite) TestLogin_CustomModel() {
	type CustomUser struct {
		ID   uint `gorm:"primaryKey"`
		Name string
	}

	s.mockConfig.EXPECT().GetString("jwt.secret").Return("Goravel").Once()
	s.mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(2).Once()

	var user CustomUser
	user.ID = 1
	user.Name = "Goravel"
	err := s.auth.Login(&user)
	s.Nil(err)
}

func (s *AuthTestSuite) TestLogin_ErrorModel() {
	type ErrorUser struct {
		ID   uint
		Name string
	}

	var errorUser ErrorUser
	errorUser.ID = 1
	errorUser.Name = "Goravel"
	err := s.auth.Login(&errorUser)
	s.EqualError(err, errors.AuthNoPrimaryKeyField.Error())
}

func (s *AuthTestSuite) TestLogin_NoPrimaryKey() {
	type User struct {
		ID   uint
		Name string
	}

	var user User
	user.ID = 1
	user.Name = "Goravel"
	err := s.auth.Login(&user)
	s.ErrorIs(err, errors.AuthNoPrimaryKeyField)
}

func (s *AuthTestSuite) TestParse_TokenDisabled() {
	token := "1"
	s.mockCache.EXPECT().GetBool("jwt:disabled:"+token, false).Return(true).Once()

	guard, err := s.auth.GetGuard("user")
	s.Nil(err)

	if guard, ok := guard.(*JwtGuard); ok {
		payload, err := guard.Parse(token)
		s.Nil(payload)
		s.EqualError(err, errors.AuthTokenDisabled.Error())
	}
}

func (s *AuthTestSuite) TestParse_TokenInvalid() {

	s.mockConfig.EXPECT().GetString("jwt.secret").Return("Goravel").Once()

	token := "1"
	s.mockCache.EXPECT().GetBool("jwt:disabled:"+token, false).Return(false).Once()

	guard, err := s.auth.GetGuard("user")
	s.Nil(err)

	if guard, ok := guard.(*JwtGuard); ok {
		payload, err := guard.Parse(token)
		s.Nil(payload)
		s.NotNil(err)
	}
}

func (s *AuthTestSuite) TestParse_TokenExpired() {

	s.mockConfig.EXPECT().GetString("jwt.secret").Return("Goravel").Twice()
	s.mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(2).Once()

	now := carbon.Now()
	issuedAt := now.StdTime()
	expireAt := now.AddMinutes(2).StdTime()

	guard, err := s.auth.GetGuard("user")
	s.Nil(err)

	err = guard.LoginUsingID(1)
	s.Nil(err)

	carbon.SetTestNow(now.AddMinutes(2))

	guardInfo, err := guard.(*JwtGuard).GetAuthToken()
	s.Nil(err)

	s.mockCache.EXPECT().GetBool("jwt:disabled:"+guardInfo.Token, false).Return(false).Once()

	if guard, ok := guard.(*JwtGuard); ok {
		payload, err := guard.Parse(guardInfo.Token)
		s.Equal(&contractsauth.Payload{
			Guard:    testUserGuard,
			Key:      "1",
			ExpireAt: jwt.NewNumericDate(expireAt).Local(),
			IssuedAt: jwt.NewNumericDate(issuedAt).Local(),
		}, payload)
		s.ErrorIs(err, errors.AuthTokenExpired)
	}

	carbon.UnsetTestNow()
}

func (s *AuthTestSuite) TestParse_InvalidCache() {
	s.mockConfig.EXPECT().GetString("auth.guards.user.driver").Return("jwt").Once()
	s.mockConfig.EXPECT().GetString("auth.guards.user.provider").Return("user").Once()
	s.mockConfig.EXPECT().GetString("auth.providers.user.driver").Return("orm").Once()
	auth, err := NewAuth(testUserGuard, nil, s.mockConfig, s.mockContext, s.mockOrm)
	s.Nil(err)

	guard, err := auth.GetGuard("user")
	s.Nil(err)

	if guard, ok := guard.(*JwtGuard); ok {
		payload, err := guard.Parse("1")
		s.Nil(payload)
		s.EqualError(err, errors.CacheSupportRequired.SetModule(errors.ModuleAuth).Error())
	}
}

func (s *AuthTestSuite) TestParse_Success() {
	s.mockConfig.EXPECT().GetString("jwt.secret").Return("Goravel").Twice()
	s.mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(2).Once()

	guard, err := s.auth.GetGuard("user")
	s.Nil(err)

	err = guard.LoginUsingID(1)
	s.Nil(err)

	if guard, ok := guard.(*JwtGuard); ok {
		guardInfo, err := guard.GetAuthToken()
		s.Nil(err)

		s.mockCache.EXPECT().GetBool("jwt:disabled:"+guardInfo.Token, false).Return(false).Once()

		payload, err := guard.Parse(guardInfo.Token)
		now := carbon.Now()
		s.Equal(&contractsauth.Payload{
			Guard:    testUserGuard,
			Key:      "1",
			ExpireAt: jwt.NewNumericDate(now.AddMinutes(2).StdTime()).Local(),
			IssuedAt: jwt.NewNumericDate(now.StdTime()).Local(),
		}, payload)
		s.Nil(err)
	}
}

func (s *AuthTestSuite) TestParse_SuccessWithPrefix() {
	carbon.SetTestNow(carbon.Now())
	s.mockConfig.EXPECT().GetString("jwt.secret").Return("Goravel").Twice()
	s.mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(2).Once()

	err := s.auth.LoginUsingID(1)
	s.Nil(err)

	guard, err := s.auth.GetGuard("user")
	s.Nil(err)

	if guard, ok := guard.(*JwtGuard); ok {
		guardInfo, err := guard.GetAuthToken()
		s.Nil(err)

		s.mockCache.EXPECT().GetBool("jwt:disabled:"+guardInfo.Token, false).Return(false).Once()

		payload, err := guard.Parse(guardInfo.Token)
		now := carbon.Now()
		s.Equal(&contractsauth.Payload{
			Guard:    testUserGuard,
			Key:      "1",
			ExpireAt: jwt.NewNumericDate(now.AddMinutes(2).StdTime()).Local(),
			IssuedAt: jwt.NewNumericDate(now.StdTime()).Local(),
		}, payload)
		s.Nil(err)
	}

	carbon.UnsetTestNow()
}

func (s *AuthTestSuite) TestParse_ExpiredAndInvalid() {
	s.mockConfig.EXPECT().GetString("jwt.secret").Return("Goravel").Once()

	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJrZXkiOiIxIiwic3ViIjoidXNlciIsImV4cCI6MTY4OTk3MDE3MiwiaWF0IjoxNjg5OTY2NTcyfQ.GApXNbicqzjF2jHsSCJ1AdziHnI1grPuJ5ddSQjGJUQ"

	s.mockCache.EXPECT().GetBool("jwt:disabled:"+token, false).Return(false).Once()

	guard, err := s.auth.GetGuard("user")
	s.Nil(err)

	if guard, ok := guard.(*JwtGuard); ok {
		_, err := guard.Parse(token)
		s.ErrorIs(err, errors.AuthInvalidToken)
	}
}

func (s *AuthTestSuite) TestUser_NoParse() {
	var user User
	err := s.auth.User(user)
	s.EqualError(err, errors.AuthParseTokenFirst.Error())
}

func (s *AuthTestSuite) TestID_NoParse() {
	// Attempt to get the ID without parsing the token first
	id, _ := s.auth.ID()
	s.Empty(id)
}

func (s *AuthTestSuite) GetGuard(name string) *JwtGuard {
	guard, err := s.auth.GetGuard(name)

	if err != nil {
		return nil
	}

	if guard, ok := guard.(*JwtGuard); ok {
		return guard
	}

	return nil
}

func (s *AuthTestSuite) TestID_Success() {
	s.mockConfig.EXPECT().GetString("jwt.secret").Return("Goravel").Twice()
	s.mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(2).Once()

	guard := s.GetGuard("user")
	s.NotNil(guard)
	err := guard.LoginUsingID(1)
	s.Nil(err)

	guardInfo, err := guard.GetAuthToken()
	s.Nil(err)

	s.mockCache.EXPECT().GetBool("jwt:disabled:"+guardInfo.Token, false).Return(false).Once()

	// Parse the token
	payload, err := guard.Parse(guardInfo.Token)
	s.Nil(err)
	s.NotNil(payload)

	// Now, call the ID method and expect it to return the correct ID
	id, _ := s.auth.ID()
	s.Equal("1", id)
}

func (s *AuthTestSuite) TestID_TokenExpired() {
	s.mockConfig.EXPECT().GetString("jwt.secret").Return("Goravel").Twice()
	s.mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(2).Once()

	guard := s.GetGuard("user")
	s.NotNil(guard)
	// Log in to get a token
	err := guard.LoginUsingID(1)
	s.Nil(err)

	guardInfo, err := guard.GetAuthToken()
	s.Nil(err)

	// Set the token as expired
	carbon.SetTestNow(carbon.Now().AddMinutes(3))

	s.mockCache.EXPECT().GetBool("jwt:disabled:"+guardInfo.Token, false).Return(false).Once()

	// Parse the token
	_, err = guard.Parse(guardInfo.Token)
	s.ErrorIs(err, errors.AuthTokenExpired)

	// Now, call the ID method and expect it to return an empty value
	id, _ := s.auth.ID()
	s.Empty(id)

	carbon.UnsetTestNow()
}

func (s *AuthTestSuite) TestID_TokenInvalid() {
	// Simulate an invalid token scenario
	s.mockConfig.EXPECT().GetString("jwt.secret").Return("Goravel").Once()

	token := "invalidToken"
	s.mockCache.EXPECT().GetBool("jwt:disabled:"+token, false).Return(false).Once()

	guard := s.GetGuard("user")
	s.NotNil(guard)

	_, err := guard.Parse(token)
	s.ErrorIs(err, errors.AuthInvalidToken)

	id, _ := s.auth.ID()
	s.Empty(id)
}

func (s *AuthTestSuite) TestUser_DBError() {
	s.mockConfig.EXPECT().GetString("jwt.secret").Return("Goravel").Twice()
	s.mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(2).Once()

	guard := s.GetGuard("user")
	s.NotNil(guard)

	err := guard.LoginUsingID(1)
	s.Nil(err)

	guradInfo, err := guard.GetAuthToken()
	s.Nil(err)

	s.mockCache.EXPECT().GetBool("jwt:disabled:"+guradInfo.Token, false).Return(false).Once()

	payload, err := guard.Parse(guradInfo.Token)
	s.NotNil(payload)
	s.Nil(err)

	var user User

	s.mockOrm.EXPECT().Query().Return(s.mockDB)
	s.mockDB.EXPECT().FindOrFail(&user, clause.Eq{Column: clause.PrimaryColumn, Value: "1"}).Return(errors.New("error")).Once()

	err = s.auth.User(&user)
	s.EqualError(err, "error")
}

func (s *AuthTestSuite) TestUser_Expired() {
	s.mockConfig.EXPECT().GetString("jwt.secret").Return("Goravel").Times(3)
	s.mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(2).Twice()

	guard := s.GetGuard("user")
	s.NotNil(guard)

	err := guard.LoginUsingID(1)
	s.Nil(err)

	guardInfo, err := guard.GetAuthToken()
	s.Nil(err)

	s.mockCache.EXPECT().GetBool("jwt:disabled:"+guardInfo.Token, false).Return(false).Once()

	carbon.SetTestNow(carbon.Now().AddMinutes(2))

	payload, err := guard.Parse(guardInfo.Token)
	s.NotNil(payload)
	s.ErrorIs(err, errors.AuthTokenExpired)

	var user User
	err = s.auth.User(&user)
	s.EqualError(err, errors.AuthTokenExpired.Error())

	s.mockConfig.EXPECT().GetInt("jwt.refresh_ttl").Return(2).Once()

	token, err := s.GetGuard("user").Refresh()
	s.NotEmpty(token)
	s.Nil(err)

	s.mockOrm.EXPECT().Query().Return(s.mockDB)
	s.mockDB.EXPECT().FindOrFail(&user, clause.Eq{Column: clause.PrimaryColumn, Value: "1"}).Return(nil).Once()

	err = s.auth.User(&user)
	s.Nil(err)

	carbon.UnsetTestNow()
}

func (s *AuthTestSuite) TestUser_RefreshExpired() {
	s.mockConfig.EXPECT().GetString("jwt.secret").Return("Goravel").Twice()
	s.mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(2).Once()

	err := s.auth.LoginUsingID(1)
	s.Nil(err)

	guard := s.GetGuard("user")
	s.NotNil(guard)

	guardInfo, err := guard.GetAuthToken()
	s.Nil(err)

	s.mockCache.EXPECT().GetBool("jwt:disabled:"+guardInfo.Token, false).Return(false).Once()

	carbon.SetTestNow(carbon.Now().AddMinutes(2))

	payload, err := guard.Parse(guardInfo.Token)
	s.NotNil(payload)
	s.ErrorIs(err, errors.AuthTokenExpired)

	var user User
	err = s.auth.User(&user)
	s.EqualError(err, errors.AuthTokenExpired.Error())

	s.mockConfig.EXPECT().GetInt("jwt.refresh_ttl").Return(1).Once()

	carbon.SetTestNow(carbon.Now().AddMinutes(2))

	token, err := guard.Refresh()
	s.Empty(token)
	s.EqualError(err, errors.AuthRefreshTimeExceeded.Error())

	carbon.UnsetTestNow()
}

func (s *AuthTestSuite) TestUser_Success() {
	s.mockConfig.EXPECT().GetString("jwt.secret").Return("Goravel").Twice()
	s.mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(2).Once()

	err := s.auth.LoginUsingID(1)
	s.Nil(err)

	guard := s.GetGuard("user")
	s.NotNil(guard)

	guardInfo, err := guard.GetAuthToken()
	s.Nil(err)

	s.mockCache.EXPECT().GetBool("jwt:disabled:"+guardInfo.Token, false).Return(false).Once()

	payload, err := guard.Parse(guardInfo.Token)
	s.NotNil(payload)
	s.Nil(err)

	var user User
	s.mockOrm.EXPECT().Query().Return(s.mockDB)
	s.mockDB.EXPECT().FindOrFail(&user, clause.Eq{Column: clause.PrimaryColumn, Value: "1"}).Return(nil).Once()

	err = guard.User(&user)
	s.Nil(err)
}

func (s *AuthTestSuite) TestUser_Success_MultipleParse() {
	testAdminGuard := "admin"

	s.mockConfig.EXPECT().GetString("jwt.secret").Return("Goravel").Times(4)
	s.mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(2)
	s.mockConfig.EXPECT().Get("auth.guards.admin.ttl").Return(2)

	s.mockConfig.EXPECT().GetString("auth.guards.admin.driver").Return("jwt").Once()
	s.mockConfig.EXPECT().GetString("auth.guards.admin.provider").Return("admin").Once()
	s.mockConfig.EXPECT().GetString("auth.providers.admin.driver").Return("orm").Once()

	guard := s.GetGuard(testUserGuard)
	s.NotNil(guard)

	err := guard.LoginUsingID(1)
	s.Nil(err)

	guard1Info, err := guard.GetAuthToken()
	s.Nil(err)

	adminGuard := s.GetGuard(testAdminGuard)
	err = adminGuard.LoginUsingID(2)
	s.Nil(err)

	guard2Info, err := adminGuard.GetAuthToken()
	s.Nil(err)

	s.mockCache.EXPECT().GetBool("jwt:disabled:"+guard1Info.Token, false).Return(false).Once()

	payload, err := guard.Parse(guard1Info.Token)
	s.Nil(err)
	s.NotNil(payload)
	s.Equal(testUserGuard, payload.Guard)
	s.Equal("1", payload.Key)

	s.mockCache.EXPECT().GetBool("jwt:disabled:"+guard2Info.Token, false).Return(false).Once()

	payload, err = adminGuard.Parse(guard2Info.Token)
	s.Nil(err)
	s.NotNil(payload)
	s.Equal(testAdminGuard, payload.Guard)
	s.Equal("2", payload.Key)

	var user1 User
	s.mockOrm.EXPECT().Query().Return(s.mockDB)
	s.mockDB.EXPECT().FindOrFail(&user1, clause.Eq{Column: clause.PrimaryColumn, Value: "1"}).Return(nil).Once()

	err = guard.User(&user1)
	s.Nil(err)

	var user2 User
	s.mockOrm.EXPECT().Query().Return(s.mockDB)
	s.mockDB.EXPECT().FindOrFail(&user2, clause.Eq{Column: clause.PrimaryColumn, Value: "2"}).Return(nil).Once()

	err = adminGuard.User(&user2)
	s.Nil(err)
}

func (s *AuthTestSuite) TestRefresh_NotParse() {
	token, err := s.GetGuard("user").Refresh()
	s.Empty(token)
	s.EqualError(err, errors.AuthParseTokenFirst.Error())
}

func (s *AuthTestSuite) TestRefresh_RefreshTimeExceeded() {
	s.mockConfig.EXPECT().GetString("jwt.secret").Return("Goravel").Twice()
	s.mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(2).Once()

	err := s.auth.LoginUsingID(2)
	s.Nil(err)

	guard := s.GetGuard("user")
	s.NotNil(guard)

	guradInfo, err := guard.GetAuthToken()
	s.Nil(err)

	s.mockCache.EXPECT().GetBool("jwt:disabled:"+guradInfo.Token, false).Return(false).Once()

	payload, err := guard.Parse(guradInfo.Token)
	s.NotNil(payload)
	s.Nil(err)

	s.mockConfig.EXPECT().GetInt("jwt.refresh_ttl").Return(1).Once()

	carbon.SetTestNow(carbon.Now().AddMinutes(4))

	token, err := guard.Refresh()
	s.Empty(token)
	s.EqualError(err, errors.AuthRefreshTimeExceeded.Error())

	carbon.UnsetTestNow()
}

func (s *AuthTestSuite) TestRefresh_Success() {
	s.mockConfig.EXPECT().GetString("jwt.secret").Return("Goravel").Times(4)
	s.mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(2).Times(3)

	err := s.auth.LoginUsingID(1)
	s.Nil(err)

	guard := s.GetGuard("user")
	s.NotNil(guard)

	guradInfo, err := guard.GetAuthToken()
	s.Nil(err)

	s.mockCache.EXPECT().GetBool("jwt:disabled:"+guradInfo.Token, false).Return(false).Once()

	payload, err := guard.Parse(guradInfo.Token)
	s.NotNil(payload)
	s.Nil(err)

	// jwt.refresh_ttl > 0
	s.mockConfig.EXPECT().GetInt("jwt.refresh_ttl").Return(1).Once()

	carbon.SetTestNow(carbon.Now().AddMinutes(2))

	token, err := guard.Refresh()
	s.NotEmpty(token)
	s.Nil(err)

	// jwt.refresh_ttl == 0
	s.mockConfig.EXPECT().GetInt("jwt.refresh_ttl").Return(0).Once()

	carbon.SetTestNow(carbon.Now().AddMinutes(2))

	token, err = guard.Refresh()
	s.NotEmpty(token)
	s.Nil(err)

	carbon.UnsetTestNow()
}

func (s *AuthTestSuite) TestLogout_CacheUnsupported() {
	s.mockConfig.EXPECT().GetString("auth.guards.user.driver").Return("jwt").Once()
	s.mockConfig.EXPECT().GetString("auth.guards.user.provider").Return("user").Once()
	s.mockConfig.EXPECT().GetString("auth.providers.user.driver").Return("orm").Once()
	if auth, err := NewAuth(testUserGuard, nil, s.mockConfig, s.mockContext, s.mockOrm); err == nil {
		s.auth = auth
	}
	s.mockConfig.EXPECT().GetString("jwt.secret").Return("Goravel").Once()
	s.mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(2).Once()

	err := s.auth.LoginUsingID(1)
	s.Nil(err)
	s.EqualError(s.auth.Logout(), errors.CacheSupportRequired.SetModule(errors.ModuleAuth).Error())
}

func (s *AuthTestSuite) TestLogout_NotParse() {
	s.EqualError(s.auth.Logout(), errors.AuthParseTokenFirst.Error())
	s.mockContext.WithValue(ctxKey, Guards{"user": nil})
	s.EqualError(s.auth.Logout(), errors.AuthParseTokenFirst.Error())
	s.mockContext.WithValue(ctxKey, Guards{"user": &AuthToken{}})
	s.EqualError(s.auth.Logout(), errors.AuthParseTokenFirst.Error())
}

func (s *AuthTestSuite) TestLogout_SetDisabledCacheError() {
	s.mockConfig.EXPECT().GetString("jwt.secret").Return("Goravel").Twice()
	s.mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(2).Twice()

	err := s.auth.LoginUsingID(1)
	s.Nil(err)

	guard := s.GetGuard("user")
	s.NotNil(guard)

	guardInfo, err := guard.GetAuthToken()
	s.Nil(err)

	s.mockCache.EXPECT().GetBool("jwt:disabled:"+guardInfo.Token, false).Return(false).Once()

	payload, err := guard.Parse(guardInfo.Token)
	s.NotNil(payload)
	s.Nil(err)

	s.mockCache.EXPECT().Put(testifymock.Anything, true, 2*time.Minute).Return(errors.New("error")).Once()

	s.EqualError(s.auth.Logout(), "error")
}

func (s *AuthTestSuite) TestLogout_Success() {
	s.mockConfig.EXPECT().GetString("jwt.secret").Return("Goravel").Twice()
	s.mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(2).Twice()

	err := s.auth.LoginUsingID(1)
	s.Nil(err)

	guard := s.GetGuard("user")
	s.NotNil(guard)

	guardInfo, err := guard.GetAuthToken()
	s.Nil(err)

	s.mockCache.EXPECT().GetBool("jwt:disabled:"+guardInfo.Token, false).Return(false).Once()

	payload, err := guard.Parse(guardInfo.Token)
	s.NotNil(payload)
	s.Nil(err)

	s.mockCache.EXPECT().Put(testifymock.Anything, true, 2*time.Minute).Return(nil).Once()

	s.Nil(s.auth.Logout())
}

func (s *AuthTestSuite) TestLogout_Success_TTL_Is_0() {
	s.mockConfig.EXPECT().GetString("jwt.secret").Return("Goravel").Twice()
	s.mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(0).Twice()

	err := s.auth.LoginUsingID(1)
	s.Nil(err)

	guard := s.GetGuard("user")
	s.NotNil(guard)

	guardInfo, err := guard.GetAuthToken()
	s.Nil(err)

	s.mockCache.EXPECT().GetBool("jwt:disabled:"+guardInfo.Token, false).Return(false).Once()

	payload, err := guard.Parse(guardInfo.Token)
	s.NotNil(payload)
	s.Nil(err)

	s.mockCache.EXPECT().Put(testifymock.Anything, true, time.Duration(60*24*365*100)*time.Minute).Return(nil).Once()

	s.Nil(s.auth.Logout())
}

func (s *AuthTestSuite) TestLogout_Error_TTL_Is_0() {
	s.mockConfig.EXPECT().GetString("jwt.secret").Return("Goravel").Twice()
	s.mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(0).Twice()

	err := s.auth.LoginUsingID(1)
	s.Nil(err)

	guard := s.GetGuard("user")
	s.NotNil(guard)

	guardInfo, err := guard.GetAuthToken()
	s.Nil(err)

	s.mockCache.EXPECT().GetBool("jwt:disabled:"+guardInfo.Token, false).Return(false).Once()

	payload, err := guard.Parse(guardInfo.Token)
	s.NotNil(payload)
	s.Nil(err)

	s.mockCache.EXPECT().Put(testifymock.Anything, true, time.Duration(60*24*365*100)*time.Minute).Return(assert.AnError).Once()

	s.EqualError(s.auth.Logout(), assert.AnError.Error())
}

func (s *AuthTestSuite) TestMakeAuthContext() {
	testAdminGuard := "admin"

	s.mockConfig.EXPECT().GetString("auth.guards.admin.driver").Return("jwt").Once()
	s.mockConfig.EXPECT().GetString("auth.guards.admin.provider").Return("admin").Once()
	s.mockConfig.EXPECT().GetString("auth.providers.admin.driver").Return("orm").Once()

	s.GetGuard("user").makeAuthContext(nil, "1")
	guards, ok := s.auth.ctx.Value(ctxKey).(Guards)
	s.True(ok)
	s.Equal(&AuthToken{nil, "1"}, guards[testUserGuard])

	s.GetGuard(testAdminGuard).makeAuthContext(nil, "2")
	guards, ok = s.auth.ctx.Value(ctxKey).(Guards)
	s.True(ok)
	s.Equal(&AuthToken{nil, "1"}, guards[testUserGuard])
	s.Equal(&AuthToken{nil, "2"}, guards[testAdminGuard])
}

func (s *AuthTestSuite) TestCheck() {
	s.mockConfig.EXPECT().GetString("jwt.secret").Return("Goravel").Once()
	s.mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(0).Once()
	s.False(s.auth.Check())
	s.True(s.auth.Guest())
	err := s.auth.LoginUsingID(1)
	s.Nil(err)
	s.True(s.auth.Check())
	s.False(s.auth.Guest())
}

func (s *AuthTestSuite) TestAuth_ExtendGuard() {
	s.auth.Extend("session", func(name string, a contractsauth.Auth, up contractsauth.UserProvider) contractsauth.Guard {
		mockGuard := mocksauth.NewGuard(s.T())
		mockGuard.EXPECT().ID().Return("session-id-xxxx", nil)
		return mockGuard
	})

	s.mockConfig.EXPECT().GetString("auth.guards.admin.driver").Return("session").Once()
	s.mockConfig.EXPECT().GetString("auth.guards.admin.provider").Return("admin").Once()
	s.mockConfig.EXPECT().GetString("auth.providers.admin.driver").Return("orm").Once()

	guard, err := s.auth.GetGuard("admin")
	s.Nil(err)

	id, err := guard.ID()
	s.Nil(err)

	s.Equal("session-id-xxxx", id)
}

func (s *AuthTestSuite) TestAuth_ExtendProvider() {
	user := User{}
	mockProvider := mocksauth.NewUserProvider(s.T())
	mockProvider.EXPECT().RetriveById(&user, "1").Return(nil).Run(func(user, id interface{}) {
		if user, ok := user.(*User); ok {
			user.Name = "MockUser"
			user.ID = 1
		}
	})

	s.auth.Provider("mock", func(auth contractsauth.Auth) (contractsauth.UserProvider, error) {
		return mockProvider, nil
	})

	s.mockConfig.EXPECT().GetString("jwt.secret").Return("Goravel").Once()
	s.mockConfig.EXPECT().GetString("auth.guards.admin.driver").Return("jwt").Once()
	s.mockConfig.EXPECT().GetString("auth.guards.admin.provider").Return("admin").Once()
	s.mockConfig.EXPECT().Get("auth.guards.admin.ttl").Return(2).Once()
	s.mockConfig.EXPECT().GetString("auth.providers.admin.driver").Return("mock").Once()

	guard, err := s.auth.GetGuard("admin")
	s.Nil(err)

	authUser := User{}
	err = guard.LoginUsingID(1)
	s.Nil(err)

	err = guard.User(&authUser)
	s.Nil(err)
	s.Equal("MockUser", authUser.Name)
	s.Equal(uint(1), authUser.ID)
}

func (s *AuthTestSuite) TestGetTtl() {
	tests := []struct {
		name     string
		setup    func()
		expected int
	}{
		{
			name: "GuardTtlIsNil",
			setup: func() {
				s.mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(nil).Once()
				s.mockConfig.EXPECT().GetInt("jwt.ttl").Return(2).Once()
			},
			expected: 2,
		},
		{
			name: "GuardTtlIsNotNil",
			setup: func() {
				s.mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(1).Once()
			},
			expected: 1,
		},
		{
			name: "GuardTtlIsZero",
			setup: func() {
				s.mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(0).Once()
			},
			expected: 60 * 24 * 365 * 100,
		},
		{
			name: "JwtTtlIsZero",
			setup: func() {
				s.mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(nil).Once()
				s.mockConfig.EXPECT().GetInt("jwt.ttl").Return(0).Once()
			},
			expected: 60 * 24 * 365 * 100,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			test.setup()
			ttl := s.GetGuard("user").getTtl()
			s.Equal(test.expected, ttl)
		})
	}
}
