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
	cachemock "github.com/goravel/framework/contracts/cache/mocks"
	configmock "github.com/goravel/framework/contracts/config/mocks"
	ormmock "github.com/goravel/framework/contracts/database/orm/mocks"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/support/carbon"
)

var guard = "user"

type User struct {
	orm.Model
	Name string
}

type Context struct {
	ctx      context.Context
	request  http.ContextRequest
	response http.ContextResponse
	values   map[string]any
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

func (mc *Context) WithValue(key string, value any) {
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
		values:   make(map[string]any),
	}
}

type AuthTestSuite struct {
	suite.Suite
	auth       *Auth
	mockCache  *cachemock.Cache
	mockConfig *configmock.Config
	mockOrm    *ormmock.Orm
	mockDB     *ormmock.Query
}

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}

func (s *AuthTestSuite) SetupTest() {
	s.mockCache = &cachemock.Cache{}
	s.mockConfig = &configmock.Config{}
	s.mockOrm = &ormmock.Orm{}
	s.mockDB = &ormmock.Query{}
	s.auth = NewAuth(guard, s.mockCache, s.mockConfig, s.mockOrm)
}

func (s *AuthTestSuite) TestLoginUsingID_EmptySecret() {
	s.mockConfig.On("GetString", "jwt.secret").Return("").Once()

	token, err := s.auth.LoginUsingID(Background(), 1)
	s.Empty(token)
	s.ErrorIs(err, ErrorEmptySecret)

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLoginUsingID_InvalidKey() {
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Once()
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	token, err := s.auth.LoginUsingID(Background(), "")
	s.Empty(token)
	s.ErrorIs(err, ErrorInvalidKey)

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLoginUsingID() {
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()

	// jwt.ttl > 0
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	token, err := s.auth.LoginUsingID(Background(), 1)
	s.NotEmpty(token)
	s.Nil(err)

	// jwt.ttl == 0
	s.mockConfig.On("GetInt", "jwt.ttl").Return(0).Once()

	token, err = s.auth.LoginUsingID(Background(), 1)
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
	token, err := s.auth.Login(Background(), &user)
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
	token, err := s.auth.Login(Background(), &user)
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
	token, err := s.auth.Login(Background(), &errorUser)
	s.Empty(token)
	s.EqualError(err, "the primaryKey field was not found in the model, set primaryKey like orm.Model")
}

func (s *AuthTestSuite) TestLogin_NoPrimaryKey() {
	type User struct {
		ID   uint
		Name string
	}

	ctx := Background()
	var user User
	user.ID = 1
	user.Name = "Goravel"
	token, err := s.auth.Login(ctx, &user)
	s.Empty(token)
	s.ErrorIs(err, ErrorNoPrimaryKeyField)
}

func (s *AuthTestSuite) TestParse_TokenDisabled() {
	token := "1"
	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(true).Once()

	payload, err := s.auth.Parse(Background(), token)
	s.Nil(payload)
	s.EqualError(err, "token is disabled")
}

func (s *AuthTestSuite) TestParse_TokenInvalid() {

	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Once()

	token := "1"
	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse(Background(), token)
	s.Nil(payload)
	s.NotNil(err)

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestParse_TokenExpired() {

	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	ctx := Background()

	now := carbon.Now()
	issuedAt := now.ToStdTime()
	expireAt := now.AddMinutes(2).ToStdTime()
	token, err := s.auth.LoginUsingID(ctx, 1)
	s.Nil(err)

	carbon.SetTestNow(now.AddMinutes(2))

	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse(ctx, token)
	s.Equal(&authcontract.Payload{
		Guard:    guard,
		Key:      "1",
		ExpireAt: jwt.NewNumericDate(expireAt).Local(),
		IssuedAt: jwt.NewNumericDate(issuedAt).Local(),
	}, payload)
	s.ErrorIs(err, ErrorTokenExpired)

	carbon.UnsetTestNow()

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestParse_InvalidCache() {
	auth := NewAuth(guard, nil, s.mockConfig, s.mockOrm)
	ctx := Background()
	payload, err := auth.Parse(ctx, "1")
	s.Nil(payload)
	s.EqualError(err, "cache support is required")
}

func (s *AuthTestSuite) TestParse_Success() {
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	ctx := Background()
	token, err := s.auth.LoginUsingID(ctx, 1)
	s.Nil(err)

	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse(ctx, token)
	s.Equal(&authcontract.Payload{
		Guard:    guard,
		Key:      "1",
		ExpireAt: jwt.NewNumericDate(carbon.Now().AddMinutes(2).ToStdTime()).Local(),
		IssuedAt: jwt.NewNumericDate(carbon.Now().ToStdTime()).Local(),
	}, payload)
	s.Nil(err)

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestParse_SuccessWithPrefix() {
	carbon.SetTestNow(carbon.Now())
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	ctx := Background()
	token, err := s.auth.LoginUsingID(ctx, 1)
	s.Nil(err)

	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse(ctx, "Bearer "+token)
	s.Equal(&authcontract.Payload{
		Guard:    guard,
		Key:      "1",
		ExpireAt: jwt.NewNumericDate(carbon.Now().AddMinutes(2).ToStdTime()).Local(),
		IssuedAt: jwt.NewNumericDate(carbon.Now().ToStdTime()).Local(),
	}, payload)
	s.Nil(err)

	carbon.UnsetTestNow()
	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestParse_ExpiredAndInvalid() {
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Once()

	ctx := Background()
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJrZXkiOiIxIiwic3ViIjoidXNlciIsImV4cCI6MTY4OTk3MDE3MiwiaWF0IjoxNjg5OTY2NTcyfQ.GApXNbicqzjF2jHsSCJ1AdziHnI1grPuJ5ddSQjGJUQ"

	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	_, err := s.auth.Parse(ctx, token)
	s.ErrorIs(err, ErrorInvalidToken)

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestUser_NoParse() {
	ctx := Background()
	var user User
	err := s.auth.User(ctx, user)
	s.EqualError(err, "parse token first")

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestUser_DBError() {
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	ctx := Background()
	token, err := s.auth.LoginUsingID(ctx, 1)
	s.Nil(err)

	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse(ctx, token)
	s.NotNil(payload)
	s.Nil(err)

	var user User

	s.mockOrm.On("Query").Return(s.mockDB)
	s.mockDB.On("FindOrFail", &user, clause.Eq{Column: clause.PrimaryColumn, Value: "1"}).Return(errors.New("error")).Once()

	err = s.auth.User(ctx, &user)
	s.EqualError(err, "error")

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestUser_Expired() {
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Times(3)
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Twice()

	ctx := Background()
	token, err := s.auth.LoginUsingID(ctx, 1)
	s.NotEmpty(token)
	s.Nil(err)

	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	carbon.SetTestNow(carbon.Now().AddMinutes(2))

	payload, err := s.auth.Parse(ctx, token)
	s.NotNil(payload)
	s.ErrorIs(err, ErrorTokenExpired)

	var user User
	err = s.auth.User(ctx, &user)
	s.EqualError(err, "token expired")

	s.mockConfig.On("GetInt", "jwt.refresh_ttl").Return(2).Once()

	token, err = s.auth.Refresh(ctx)
	s.NotEmpty(token)
	s.Nil(err)

	s.mockOrm.On("Query").Return(s.mockDB)
	s.mockDB.On("FindOrFail", &user, clause.Eq{Column: clause.PrimaryColumn, Value: "1"}).Return(nil).Once()

	err = s.auth.User(ctx, &user)
	s.Nil(err)

	carbon.UnsetTestNow()

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestUser_RefreshExpired() {
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	ctx := Background()
	token, err := s.auth.LoginUsingID(ctx, 1)
	s.NotEmpty(token)
	s.Nil(err)

	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	carbon.SetTestNow(carbon.Now().AddMinutes(2))

	payload, err := s.auth.Parse(ctx, token)
	s.NotNil(payload)
	s.ErrorIs(err, ErrorTokenExpired)

	var user User
	err = s.auth.User(ctx, &user)
	s.EqualError(err, "token expired")

	s.mockConfig.On("GetInt", "jwt.refresh_ttl").Return(1).Once()

	carbon.SetTestNow(carbon.Now().AddMinutes(2))

	token, err = s.auth.Refresh(ctx)
	s.Empty(token)
	s.EqualError(err, "refresh time exceeded")

	carbon.UnsetTestNow()

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestUser_Success() {
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	ctx := Background()
	token, err := s.auth.LoginUsingID(ctx, 1)
	s.Nil(err)

	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse(ctx, token)
	s.NotNil(payload)
	s.Nil(err)

	var user User
	s.mockOrm.On("Query").Return(s.mockDB)
	s.mockDB.On("FindOrFail", &user, clause.Eq{Column: clause.PrimaryColumn, Value: "1"}).Return(nil).Once()

	err = s.auth.User(ctx, &user)
	s.Nil(err)

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestRefresh_NotParse() {
	ctx := Background()
	token, err := s.auth.Refresh(ctx)
	s.Empty(token)
	s.EqualError(err, "parse token first")

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestRefresh_RefreshTimeExceeded() {
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	ctx := Background()
	token, err := s.auth.LoginUsingID(ctx, 2)

	s.Nil(err)

	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse(ctx, token)
	s.NotNil(payload)
	s.Nil(err)

	s.mockConfig.On("GetInt", "jwt.refresh_ttl").Return(1).Once()

	carbon.SetTestNow(carbon.Now().AddMinutes(4))

	token, err = s.auth.Refresh(ctx)
	s.Empty(token)
	s.EqualError(err, "refresh time exceeded")

	carbon.UnsetTestNow()

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestRefresh_Success() {
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Times(4)
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Times(3)

	ctx := Background()
	token, err := s.auth.LoginUsingID(ctx, 1)
	s.Nil(err)

	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse(ctx, token)
	s.NotNil(payload)
	s.Nil(err)

	// jwt.refresh_ttl > 0
	s.mockConfig.On("GetInt", "jwt.refresh_ttl").Return(1).Once()

	carbon.SetTestNow(carbon.Now().AddMinutes(2))

	token, err = s.auth.Refresh(ctx)
	s.NotEmpty(token)
	s.Nil(err)

	// jwt.refresh_ttl == 0
	s.mockConfig.On("GetInt", "jwt.refresh_ttl").Return(0).Once()

	carbon.SetTestNow(carbon.Now().AddMinutes(2))

	token, err = s.auth.Refresh(ctx)
	s.NotEmpty(token)
	s.Nil(err)

	carbon.UnsetTestNow()

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLogout_CacheUnsupported() {
	s.auth = NewAuth(guard, nil, s.mockConfig, s.mockOrm)
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Once()
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	ctx := Background()
	token, err := s.auth.LoginUsingID(ctx, 1)
	s.NotEmpty(token)
	s.Nil(err)
	s.EqualError(s.auth.Logout(ctx), "cache support is required")

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLogout_NotParse() {
	s.Nil(s.auth.Logout(Background()))
}

func (s *AuthTestSuite) TestLogout_SetDisabledCacheError() {
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Twice()

	ctx := Background()
	token, err := s.auth.LoginUsingID(ctx, 1)
	s.Nil(err)

	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse(ctx, token)
	s.NotNil(payload)
	s.Nil(err)

	s.mockCache.On("Put", testifymock.Anything, true, 2*time.Minute).Return(errors.New("error")).Once()

	s.EqualError(s.auth.Logout(ctx), "error")

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLogout_Success() {
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	s.mockConfig.On("GetInt", "jwt.ttl").Return(2).Twice()

	ctx := Background()
	token, err := s.auth.LoginUsingID(ctx, 1)
	s.NotEmpty(token)
	s.Nil(err)

	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse(ctx, token)
	s.NotNil(payload)
	s.Nil(err)

	s.mockCache.On("Put", testifymock.Anything, true, 2*time.Minute).Return(nil).Once()

	s.Nil(s.auth.Logout(ctx))

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLogout_Success_TTL_Is_0() {
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	s.mockConfig.On("GetInt", "jwt.ttl").Return(0).Twice()

	ctx := Background()
	token, err := s.auth.LoginUsingID(ctx, 1)
	s.NotEmpty(token)
	s.Nil(err)

	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse(ctx, token)
	s.NotNil(payload)
	s.Nil(err)

	s.mockCache.On("Forever", testifymock.Anything, true).Return(true).Once()

	s.Nil(s.auth.Logout(ctx))

	s.mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLogout_Error_TTL_Is_0() {
	s.mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	s.mockConfig.On("GetInt", "jwt.ttl").Return(0).Twice()

	ctx := Background()
	token, err := s.auth.LoginUsingID(ctx, 1)
	s.NotEmpty(token)
	s.Nil(err)

	s.mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse(ctx, token)
	s.NotNil(payload)
	s.Nil(err)

	s.mockCache.On("Forever", testifymock.Anything, true).Return(false).Once()

	s.EqualError(s.auth.Logout(ctx), "cache forever failed")

	s.mockConfig.AssertExpectations(s.T())
}
