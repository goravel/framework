package auth

import (
	"errors"
	"testing"
	"time"

	contractauth "github.com/goravel/framework/contracts/auth"
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/http"
	"github.com/goravel/framework/testing/mock"

	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

var guard = "user"

type User struct {
	orm.Model
	Name string
}

type AuthTestSuite struct {
	suite.Suite
}

var app contractauth.Auth

func TestAuthTestSuite(t *testing.T) {
	app = NewApplication(guard)

	suite.Run(t, new(AuthTestSuite))
}

func (s *AuthTestSuite) SetupTest() {
	unit = time.Second
}

func (s *AuthTestSuite) TestLoginUsingID_EmptySecret() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("").Once()

	token, err := app.LoginUsingID(http.Background(), 1)
	assert.Empty(s.T(), token)
	assert.ErrorIs(s.T(), err, ErrorEmptySecret)

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLoginUsingID() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel").Once()
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	token, err := app.LoginUsingID(http.Background(), 1)
	assert.NotEmpty(s.T(), token)
	assert.Nil(s.T(), err)

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLogin_Model() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel").Once()
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	var user User
	user.ID = 1
	user.Name = "Goravel"
	token, err := app.Login(http.Background(), &user)
	assert.NotEmpty(s.T(), token)
	assert.Nil(s.T(), err)

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLogin_CustomModel() {
	type CustomUser struct {
		ID   uint `gorm:"primaryKey"`
		Name string
	}

	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel").Once()
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	var user CustomUser
	user.ID = 1
	user.Name = "Goravel"
	token, err := app.Login(http.Background(), &user)
	assert.NotEmpty(s.T(), token)
	assert.Nil(s.T(), err)

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLogin_ErrorModel() {
	type ErrorUser struct {
		ID   uint
		Name string
	}

	var errorUser ErrorUser
	errorUser.ID = 1
	errorUser.Name = "Goravel"
	token, err := app.Login(http.Background(), &errorUser)
	assert.Empty(s.T(), token)
	assert.EqualError(s.T(), err, "the primaryKey field was not found in the model, set primaryKey like orm.Model")
}

func (s *AuthTestSuite) TestParse_TokenDisabled() {
	token := "1"
	mockCache := mock.Cache()
	mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(true).Once()

	err := app.Parse(http.Background(), token)
	assert.EqualError(s.T(), err, "token is disabled")
}

func (s *AuthTestSuite) TestParse_TokenInvalid() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel").Once()

	token := "1"
	mockCache := mock.Cache()
	mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	err := app.Parse(http.Background(), token)
	assert.NotNil(s.T(), err)

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestParse_TokenExpired() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel")
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	ctx := http.Background()
	token, err := app.LoginUsingID(ctx, 1)
	assert.Nil(s.T(), err)

	time.Sleep(2 * unit)

	mockCache := mock.Cache()
	mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	err = app.Parse(ctx, token)
	assert.ErrorIs(s.T(), err, ErrorTokenExpired)

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestParse_Success() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel")
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	ctx := http.Background()
	token, err := app.LoginUsingID(ctx, 1)
	assert.Nil(s.T(), err)

	mockCache := mock.Cache()
	mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	err = app.Parse(ctx, token)
	assert.Nil(s.T(), err)

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestParse_SuccessWithPrefix() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel")
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	ctx := http.Background()
	token, err := app.LoginUsingID(ctx, 1)
	assert.Nil(s.T(), err)

	mockCache := mock.Cache()
	mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	err = app.Parse(ctx, "Bearer "+token)
	assert.Nil(s.T(), err)

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestUser_NoParse() {
	mockConfig := mock.Config()

	ctx := http.Background()
	var user User
	err := app.User(ctx, user)
	assert.EqualError(s.T(), err, "parse token first")

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestUser_DBError() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel")
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	ctx := http.Background()
	token, err := app.LoginUsingID(ctx, 1)
	assert.Nil(s.T(), err)

	mockCache := mock.Cache()
	mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	err = app.Parse(ctx, token)
	assert.Nil(s.T(), err)

	var user User

	mockOrm, mockDB, _ := mock.Orm()
	mockOrm.On("Query").Return(mockDB)
	mockDB.On("Find", &user, "1").Return(errors.New("error")).Once()

	err = app.User(ctx, &user)
	assert.EqualError(s.T(), err, "error")

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestUser_Expired() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel")
	mockConfig.On("GetInt", "jwt.ttl").Return(2)

	ctx := http.Background()
	token, err := app.LoginUsingID(ctx, 1)
	assert.NotEmpty(s.T(), token)
	assert.Nil(s.T(), err)

	mockCache := mock.Cache()
	mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	time.Sleep(2 * unit)

	err = app.Parse(ctx, token)
	assert.ErrorIs(s.T(), err, ErrorTokenExpired)

	var user User
	err = app.User(ctx, &user)
	assert.EqualError(s.T(), err, "token expired")

	mockConfig.On("GetInt", "jwt.refresh_ttl").Return(2).Once()

	token, err = app.Refresh(ctx)
	assert.NotEmpty(s.T(), token)
	assert.Nil(s.T(), err)

	mockOrm, mockDB, _ := mock.Orm()
	mockOrm.On("Query").Return(mockDB)
	mockDB.On("Find", &user, "1").Return(nil).Once()

	err = app.User(ctx, &user)
	assert.Nil(s.T(), err)

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestUser_RefreshExpired() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel")
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	ctx := http.Background()
	token, err := app.LoginUsingID(ctx, 1)
	assert.NotEmpty(s.T(), token)
	assert.Nil(s.T(), err)

	mockCache := mock.Cache()
	mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	time.Sleep(2 * unit)

	err = app.Parse(ctx, token)
	assert.ErrorIs(s.T(), err, ErrorTokenExpired)

	var user User
	err = app.User(ctx, &user)
	assert.EqualError(s.T(), err, "token expired")

	mockConfig.On("GetInt", "jwt.refresh_ttl").Return(1).Once()

	time.Sleep(2 * unit)

	token, err = app.Refresh(ctx)
	assert.Empty(s.T(), token)
	assert.EqualError(s.T(), err, "refresh time exceeded")

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestUser_Success() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel")
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	ctx := http.Background()
	token, err := app.LoginUsingID(ctx, 1)
	assert.Nil(s.T(), err)

	mockCache := mock.Cache()
	mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	err = app.Parse(ctx, token)
	assert.Nil(s.T(), err)

	var user User
	mockOrm, mockDB, _ := mock.Orm()
	mockOrm.On("Query").Return(mockDB)
	mockDB.On("Find", &user, "1").Return(nil).Once()

	err = app.User(ctx, &user)
	assert.Nil(s.T(), err)

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestRefresh_NotParse() {
	mockConfig := mock.Config()

	ctx := http.Background()
	token, err := app.Refresh(ctx)
	assert.Empty(s.T(), token)
	assert.EqualError(s.T(), err, "parse token first")

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestRefresh_RefreshTimeExceeded() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel")
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	ctx := http.Background()
	token, err := app.LoginUsingID(ctx, 1)
	assert.Nil(s.T(), err)

	mockCache := mock.Cache()
	mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	err = app.Parse(ctx, token)
	assert.Nil(s.T(), err)

	mockConfig.On("GetInt", "jwt.refresh_ttl").Return(1).Once()
	time.Sleep(4 * unit)

	token, err = app.Refresh(ctx)
	assert.Empty(s.T(), token)
	assert.EqualError(s.T(), err, "refresh time exceeded")

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestRefresh_Success() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel")
	mockConfig.On("GetInt", "jwt.ttl").Return(2)

	ctx := http.Background()
	token, err := app.LoginUsingID(ctx, 1)
	assert.Nil(s.T(), err)

	mockCache := mock.Cache()
	mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	err = app.Parse(ctx, token)
	assert.Nil(s.T(), err)

	mockConfig.On("GetInt", "jwt.refresh_ttl").Return(1).Once()
	time.Sleep(2 * unit)

	token, err = app.Refresh(ctx)
	assert.NotEmpty(s.T(), token)
	assert.Nil(s.T(), err)

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLogout_CacheUnsupported() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel")
	mockConfig.On("GetInt", "jwt.ttl").Return(2)

	ctx := http.Background()
	token, err := app.LoginUsingID(ctx, 1)
	assert.NotEmpty(s.T(), token)
	assert.Nil(s.T(), err)
	assert.EqualError(s.T(), app.Logout(ctx), "cache support is required")

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLogout_NotParse() {
	assert.Nil(s.T(), app.Logout(http.Background()))
}

func (s *AuthTestSuite) TestLogout_SetDisabledCacheError() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel")
	mockConfig.On("GetInt", "jwt.ttl").Return(2)

	ctx := http.Background()
	token, err := app.LoginUsingID(ctx, 1)
	assert.Nil(s.T(), err)

	mockCache := mock.Cache()
	mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	err = app.Parse(ctx, token)
	assert.Nil(s.T(), err)

	mockCache.On("Put", testifymock.Anything, true, 2*unit).Return(errors.New("error")).Once()

	assert.EqualError(s.T(), app.Logout(ctx), "error")

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLogout_Success() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel")
	mockConfig.On("GetInt", "jwt.ttl").Return(2)

	ctx := http.Background()
	token, err := app.LoginUsingID(ctx, 1)
	assert.NotEmpty(s.T(), token)
	assert.Nil(s.T(), err)

	mockCache := mock.Cache()
	mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	err = app.Parse(ctx, token)
	assert.Nil(s.T(), err)

	mockCache.On("Put", testifymock.Anything, true, 2*unit).Return(nil).Once()

	assert.Nil(s.T(), app.Logout(ctx))

	mockConfig.AssertExpectations(s.T())
}
