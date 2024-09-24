package auth

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm/clause"

	authcontract "github.com/goravel/framework/contracts/auth"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/database/orm"
	cachemock "github.com/goravel/framework/mocks/cache"
	configmock "github.com/goravel/framework/mocks/config"
	ormmock "github.com/goravel/framework/mocks/database/orm"
	"github.com/goravel/framework/support/carbon"
)

var testUserGuard = "user"

type User struct {
	orm.Model
	Name string
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
	mockCache   *cachemock.Cache
	mockConfig  *configmock.Config
	mockContext http.Context
	mockOrm     *ormmock.Orm
	mockDB      *ormmock.Query
}

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}

func (s *AuthTestSuite) SetupTest() {
	s.mockCache = &cachemock.Cache{}
	s.mockConfig = &configmock.Config{}
	s.mockContext = Background()
	s.mockOrm = &ormmock.Orm{}
	s.mockDB = &ormmock.Query{}
	s.auth = NewAuth(testUserGuard, s.mockCache, s.mockConfig, s.mockContext, s.mockOrm)
}

func (s *AuthTestSuite) TestLoginUsingID_EmptySecret() {
	s.mockConfig.On("GetString", "jwt.secret").Return("").Once()

	token, err := s.auth.LoginUsingID(1)
	s.Empty(token)
	s.ErrorIs(err, ErrorEmptySecret)

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLoginUsingID_InvalidKey() {
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Once()
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	token, err := s.auth.LoginUsingID("")
	s.Empty(token)
	s.ErrorIs(err, ErrorInvalidKey)

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLoginUsingID() {
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()

	// jwt.ttl > 0
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	token, err := s.auth.LoginUsingID(1)
	s.NotEmpty(token)
	s.Nil(err)

	// jwt.ttl == 0
	s.mockConfig.On("GetInt", "jwt.ttl").Return(0).Once()

	token, err = s.auth.LoginUsingID(1)
	s.NotEmpty(token)
	s.Nil(err)

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLogin_Model() {

	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Once()
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	var user User
	user.ID = 1
	user.Name = "Goravel"
	token, err := s.auth.Login(&user)
	s.NotEmpty(token)
	s.Nil(err)

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLogin_CustomModel() {
	type CustomUser struct {
		ID   uint `gorm:"primaryKey"`
		Name string
	}

	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Once()
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	var user CustomUser
	user.ID = 1
	user.Name = "Goravel"
	token, err := s.auth.Login(&user)
	s.NotEmpty(token)
	s.Nil(err)

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLogin_ErrorModel() {
	type ErrorUser struct {
		ID   uint
		Name string
	}

	var errorUser ErrorUser
	errorUser.ID = 1
	errorUser.Name = "Goravel"
	token, err := s.auth.Login(&errorUser)
	s.Empty(token)
	s.EqualError(err, "the primaryKey field was not found in the model, set primaryKey like orm.Model")
}

func (s *AuthTestSuite) TestLogin_NoPrimaryKey() {
	type User struct {
		ID   uint
		Name string
	}

	var user User
	user.ID = 1
	user.Name = "Goravel"
	token, err := s.auth.Login(&user)
	s.Empty(token)
	s.ErrorIs(err, ErrorNoPrimaryKeyField)
}

func (s *AuthTestSuite) TestParse_TokenDisabled() {
	token := "1"
	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(true).Once()

	payload, err := s.auth.Parse(token)
	s.Nil(payload)
	s.EqualError(err, "token is disabled")
}

func (s *AuthTestSuite) TestParse_TokenInvalid() {

	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Once()

	token := "1"
	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse(token)
	s.Nil(payload)
	s.NotNil(err)

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestParse_TokenExpired() {

	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	now := carbon.Now()
	issuedAt := now.StdTime()
	expireAt := now.AddMinutes(2).StdTime()
	token, err := s.auth.LoginUsingID(1)
	s.Nil(err)

	carbon.SetTestNow(now.AddMinutes(2))

	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse(token)
	s.Equal(&authcontract.Payload{
		Guard:    testUserGuard,
		Key:      "1",
		ExpireAt: jwt.NewNumericDate(expireAt).Local(),
		IssuedAt: jwt.NewNumericDate(issuedAt).Local(),
	}, payload)
	s.ErrorIs(err, ErrorTokenExpired)

	carbon.UnsetTestNow()

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestParse_InvalidCache() {
	auth := NewAuth(testUserGuard, nil, s.mockConfig, s.mockContext, s.mockOrm)
	payload, err := auth.Parse("1")
	s.Nil(payload)
	s.EqualError(err, "cache support is required")
}

func (s *AuthTestSuite) TestParse_Success() {
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	token, err := s.auth.LoginUsingID(1)
	s.Nil(err)

	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse(token)
	s.Equal(&authcontract.Payload{
		Guard:    testUserGuard,
		Key:      "1",
		ExpireAt: jwt.NewNumericDate(carbon.Now().AddMinutes(2).StdTime()).Local(),
		IssuedAt: jwt.NewNumericDate(carbon.Now().StdTime()).Local(),
	}, payload)
	s.Nil(err)

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestParse_SuccessWithPrefix() {
	carbon.SetTestNow(carbon.Now())
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	token, err := s.auth.LoginUsingID(1)
	s.Nil(err)

	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse("Bearer " + token)
	s.Equal(&authcontract.Payload{
		Guard:    testUserGuard,
		Key:      "1",
		ExpireAt: jwt.NewNumericDate(carbon.Now().AddMinutes(2).StdTime()).Local(),
		IssuedAt: jwt.NewNumericDate(carbon.Now().StdTime()).Local(),
	}, payload)
	s.Nil(err)

	carbon.UnsetTestNow()
	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestParse_ExpiredAndInvalid() {
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Once()

	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJrZXkiOiIxIiwic3ViIjoidXNlciIsImV4cCI6MTY4OTk3MDE3MiwiaWF0IjoxNjg5OTY2NTcyfQ.GApXNbicqzjF2jHsSCJ1AdziHnI1grPuJ5ddSQjGJUQ"

	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	_, err := s.auth.Parse(token)
	s.ErrorIs(err, ErrorInvalidToken)

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestUser_NoParse() {
	var user User
	err := s.auth.User(user)
	s.EqualError(err, "parse token first")

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestID_NoParse() {
	// Attempt to get the ID without parsing the token first
	id, _ := s.auth.Id()
	s.Empty(id)
}

func (s *AuthTestSuite) TestID_Success() {
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	// Log in to get a token
	token, err := s.auth.LoginUsingID(1)
	s.Nil(err)

	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	// Parse the token
	payload, err := s.auth.Parse(token)
	s.Nil(err)
	s.NotNil(payload)

	// Now, call the ID method and expect it to return the correct ID
	id, _ := s.auth.Id()
	s.Equal("1", id)
}

func (s *AuthTestSuite) TestID_TokenExpired() {
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	// Log in to get a token
	token, err := s.auth.LoginUsingID(1)
	s.Nil(err)

	// Set the token as expired
	carbon.SetTestNow(carbon.Now().AddMinutes(3))

	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	// Parse the token
	_, err = s.auth.Parse(token)
	s.ErrorIs(err, ErrorTokenExpired)

	// Now, call the ID method and expect it to return an empty value
	id, _ := s.auth.Id()
	s.Empty(id)

	carbon.UnsetTestNow()
}

func (s *AuthTestSuite) TestID_TokenInvalid() {
	// Simulate an invalid token scenario
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Once()

	token := "invalidToken"
	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	_, err := s.auth.Parse(token)
	s.ErrorIs(err, ErrorInvalidToken)

	id, _ := s.auth.Id()
	s.Empty(id)
}

func (s *AuthTestSuite) TestUser_DBError() {
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	token, err := s.auth.LoginUsingID(1)
	s.Nil(err)

	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse(token)
	s.NotNil(payload)
	s.Nil(err)

	var user User

	s.mockOrm.On("Query").Return(s.mockDB)
	s.mockDB.On("FindOrFail", &user, clause.Eq{Column: clause.PrimaryColumn, Value: "1"}).Return(errors.New("error")).Once()

	err = s.auth.User(&user)
	s.EqualError(err, "error")

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestUser_Expired() {
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Times(3)
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Twice()

	token, err := s.auth.LoginUsingID(1)
	s.NotEmpty(token)
	s.Nil(err)

	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	carbon.SetTestNow(carbon.Now().AddMinutes(2))

	payload, err := s.auth.Parse(token)
	s.NotNil(payload)
	s.ErrorIs(err, ErrorTokenExpired)

	var user User
	err = s.auth.User(&user)
	s.EqualError(err, "token expired")

	s.mockConfig.On("GetInt", "jwt.refresh_ttl").Return(2).Once()

	token, err = s.auth.Refresh()
	s.NotEmpty(token)
	s.Nil(err)

	s.mockOrm.On("Query").Return(s.mockDB)
	s.mockDB.On("FindOrFail", &user, clause.Eq{Column: clause.PrimaryColumn, Value: "1"}).Return(nil).Once()

	err = s.auth.User(&user)
	s.Nil(err)

	carbon.UnsetTestNow()

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestUser_RefreshExpired() {
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	token, err := s.auth.LoginUsingID(1)
	s.NotEmpty(token)
	s.Nil(err)

	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	carbon.SetTestNow(carbon.Now().AddMinutes(2))

	payload, err := s.auth.Parse(token)
	s.NotNil(payload)
	s.ErrorIs(err, ErrorTokenExpired)

	var user User
	err = s.auth.User(&user)
	s.EqualError(err, "token expired")

	s.mockConfig.On("GetInt", "jwt.refresh_ttl").Return(1).Once()

	carbon.SetTestNow(carbon.Now().AddMinutes(2))

	token, err = s.auth.Refresh()
	s.Empty(token)
	s.EqualError(err, "refresh time exceeded")

	carbon.UnsetTestNow()

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestUser_Success() {
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	token, err := s.auth.LoginUsingID(1)
	s.Nil(err)

	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse(token)
	s.NotNil(payload)
	s.Nil(err)

	var user User
	s.mockOrm.On("Query").Return(s.mockDB)
	s.mockDB.On("FindOrFail", &user, clause.Eq{Column: clause.PrimaryColumn, Value: "1"}).Return(nil).Once()

	err = s.auth.User(&user)
	s.Nil(err)

	s.mockConfig.AssertExpectations(s.T())
	s.mockCache.AssertExpectations(s.T())
	s.mockOrm.AssertExpectations(s.T())
	s.mockDB.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestUser_Success_MultipleParse() {
	testAdminGuard := "admin"

	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	token1, err := s.auth.LoginUsingID(1)
	s.Nil(err)

	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	token2, err := s.auth.Guard(testAdminGuard).LoginUsingID(2)
	s.Nil(err)

	s.mockCache.On("GetBool", "jwt:disabled:"+token1, false).Return(false).Once()

	payload, err := s.auth.Parse(token1)
	s.Nil(err)
	s.NotNil(payload)
	s.Equal(testUserGuard, payload.Guard)
	s.Equal("1", payload.Key)

	s.mockCache.On("GetBool", "jwt:disabled:"+token2, false).Return(false).Once()

	payload, err = s.auth.Guard(testAdminGuard).Parse(token2)
	s.Nil(err)
	s.NotNil(payload)
	s.Equal(testAdminGuard, payload.Guard)
	s.Equal("2", payload.Key)

	var user1 User
	s.mockOrm.On("Query").Return(s.mockDB)
	s.mockDB.On("FindOrFail", &user1, clause.Eq{Column: clause.PrimaryColumn, Value: "1"}).Return(nil).Once()

	err = s.auth.User(&user1)
	s.Nil(err)

	var user2 User
	s.mockOrm.On("Query").Return(s.mockDB)
	s.mockDB.On("FindOrFail", &user2, clause.Eq{Column: clause.PrimaryColumn, Value: "2"}).Return(nil).Once()

	err = s.auth.Guard(testAdminGuard).User(&user2)
	s.Nil(err)

	s.mockConfig.AssertExpectations(s.T())
	s.mockCache.AssertExpectations(s.T())
	s.mockOrm.AssertExpectations(s.T())
	s.mockDB.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestRefresh_NotParse() {
	token, err := s.auth.Refresh()
	s.Empty(token)
	s.EqualError(err, "parse token first")

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestRefresh_RefreshTimeExceeded() {
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	token, err := s.auth.LoginUsingID(2)

	s.Nil(err)

	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse(token)
	s.NotNil(payload)
	s.Nil(err)

	s.mockConfig.On("GetInt", "jwt.refresh_ttl").Return(1).Once()

	carbon.SetTestNow(carbon.Now().AddMinutes(4))

	token, err = s.auth.Refresh()
	s.Empty(token)
	s.EqualError(err, "refresh time exceeded")

	carbon.UnsetTestNow()

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestRefresh_Success() {
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Times(4)
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Times(3)

	token, err := s.auth.LoginUsingID(1)
	s.Nil(err)

	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse(token)
	s.NotNil(payload)
	s.Nil(err)

	// jwt.refresh_ttl > 0
	s.mockConfig.On("GetInt", "jwt.refresh_ttl").Return(1).Once()

	carbon.SetTestNow(carbon.Now().AddMinutes(2))

	token, err = s.auth.Refresh()
	s.NotEmpty(token)
	s.Nil(err)

	// jwt.refresh_ttl == 0
	s.mockConfig.On("GetInt", "jwt.refresh_ttl").Return(0).Once()

	carbon.SetTestNow(carbon.Now().AddMinutes(2))

	token, err = s.auth.Refresh()
	s.NotEmpty(token)
	s.Nil(err)

	carbon.UnsetTestNow()

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLogout_CacheUnsupported() {
	s.auth = NewAuth(testUserGuard, nil, s.mockConfig, s.mockContext, s.mockOrm)
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Once()
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	token, err := s.auth.LoginUsingID(1)
	s.NotEmpty(token)
	s.Nil(err)
	s.EqualError(s.auth.Logout(), "cache support is required")

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLogout_NotParse() {
	s.Nil(s.auth.Logout())
}

func (s *AuthTestSuite) TestLogout_SetDisabledCacheError() {
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Twice()

	token, err := s.auth.LoginUsingID(1)
	s.Nil(err)

	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse(token)
	s.NotNil(payload)
	s.Nil(err)

	s.mockCache.On("Put", testifymock.Anything, true, 2*time.Minute).Return(errors.New("error")).Once()

	s.EqualError(s.auth.Logout(), "error")

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLogout_Success() {
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Twice()

	token, err := s.auth.LoginUsingID(1)
	s.NotEmpty(token)
	s.Nil(err)

	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse(token)
	s.NotNil(payload)
	s.Nil(err)

	s.mockCache.On("Put", testifymock.Anything, true, 2*time.Minute).Return(nil).Once()

	s.Nil(s.auth.Logout())

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLogout_Success_TTL_Is_0() {
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	s.mockConfig.On("GetInt", "jwt.ttl").Return(0).Twice()

	token, err := s.auth.LoginUsingID(1)
	s.NotEmpty(token)
	s.Nil(err)

	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse(token)
	s.NotNil(payload)
	s.Nil(err)

	s.mockCache.On("Forever", testifymock.Anything, true).Return(true).Once()

	s.Nil(s.auth.Logout())

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLogout_Error_TTL_Is_0() {
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	s.mockConfig.On("GetInt", "jwt.ttl").Return(0).Twice()

	token, err := s.auth.LoginUsingID(1)
	s.NotEmpty(token)
	s.Nil(err)

	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse(token)
	s.NotNil(payload)
	s.Nil(err)

	s.mockCache.On("Forever", testifymock.Anything, true).Return(false).Once()

	s.EqualError(s.auth.Logout(), "cache forever failed")

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestMakeAuthContext() {
	testAdminGuard := "admin"

	s.auth.makeAuthContext(nil, "1")
	guards, ok := s.auth.ctx.Value(ctxKey).(Guards)
	s.True(ok)
	s.Equal(&Guard{nil, "1"}, guards[testUserGuard])

	s.auth.Guard(testAdminGuard).(*Auth).makeAuthContext(nil, "2")
	guards, ok = s.auth.ctx.Value(ctxKey).(Guards)
	s.True(ok)
	s.Equal(&Guard{nil, "1"}, guards[testUserGuard])
	s.Equal(&Guard{nil, "2"}, guards[testAdminGuard])
}
