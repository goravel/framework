package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/testing/mock"

	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type User struct {
	orm.Model
	Name string
}

type AuthTestSuite struct {
	suite.Suite
}

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}

func (s *AuthTestSuite) SetupTest() {
	unit = time.Second
}

func (s *AuthTestSuite) TestLoginUsingID() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel").Once()
	mockConfig.On("GetString", "auth.defaults.guard").Return("user")
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	app := NewApplication()
	token, err := app.LoginUsingID(1)
	assert.NotEmpty(s.T(), token)
	assert.Nil(s.T(), err)

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLogin_Model() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel").Once()
	mockConfig.On("GetString", "auth.defaults.guard").Return("user")
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	app := NewApplication()

	var user User
	user.ID = 1
	user.Name = "Goravel"
	token, err := app.Login(&user)
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
	mockConfig.On("GetString", "auth.defaults.guard").Return("user")
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	app := NewApplication()

	var user CustomUser
	user.ID = 1
	user.Name = "Goravel"
	token, err := app.Login(&user)
	assert.NotEmpty(s.T(), token)
	assert.Nil(s.T(), err)

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLogin_ErrorModel() {
	type ErrorUser struct {
		ID   uint
		Name string
	}

	app := NewApplication()

	var errorUser ErrorUser
	errorUser.ID = 1
	errorUser.Name = "Goravel"
	token, err := app.Login(&errorUser)
	assert.Empty(s.T(), token)
	assert.EqualError(s.T(), err, "the primaryKey field was not found in the model, set primaryKey like orm.Model")
}

func (s *AuthTestSuite) TestParse_TokenDisabled() {
	token := "1"
	mockCache := mock.Cache()
	mockCache.On("Get", "jwt:disabled:"+token, false).Return(true).Once()

	app := NewApplication()

	expired, err := app.Parse(token)
	assert.False(s.T(), expired)
	assert.EqualError(s.T(), err, "token is disabled")
}

func (s *AuthTestSuite) TestParse_TokenInvalid() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel").Once()

	token := "1"
	mockCache := mock.Cache()
	mockCache.On("Get", "jwt:disabled:"+token, false).Return(false).Once()

	app := NewApplication()

	expired, err := app.Parse(token)
	assert.False(s.T(), expired)
	assert.NotNil(s.T(), err)

	mockConfig.AssertExpectations(s.T())

}

func (s *AuthTestSuite) TestParse_TokenExpired() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel")
	mockConfig.On("GetString", "auth.defaults.guard").Return("user")
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	app := NewApplication()
	token, err := app.LoginUsingID(1)
	assert.Nil(s.T(), err)

	time.Sleep(2 * unit)

	mockCache := mock.Cache()
	mockCache.On("Get", "jwt:disabled:"+token, false).Return(false).Once()

	expired, err := app.Parse(token)
	assert.True(s.T(), expired)
	assert.Nil(s.T(), err)

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestParse_Success() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel")
	mockConfig.On("GetString", "auth.defaults.guard").Return("user")
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	app := NewApplication()
	token, err := app.LoginUsingID(1)
	assert.Nil(s.T(), err)

	mockCache := mock.Cache()
	mockCache.On("Get", "jwt:disabled:"+token, false).Return(false).Once()

	expired, err := app.Parse(token)
	assert.False(s.T(), expired)
	assert.Nil(s.T(), err)

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestParse_SuccessWithPrefix() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel")
	mockConfig.On("GetString", "auth.defaults.guard").Return("user")
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	app := NewApplication()
	token, err := app.LoginUsingID(1)
	assert.Nil(s.T(), err)

	mockCache := mock.Cache()
	mockCache.On("Get", "jwt:disabled:"+token, false).Return(false).Once()

	expired, err := app.Parse("Bearer " + token)
	assert.False(s.T(), expired)
	assert.Nil(s.T(), err)

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestUser_NoParse() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "auth.defaults.guard").Return("user")

	app := NewApplication()

	var user User
	err := app.User(user)
	assert.EqualError(s.T(), err, "parse token first")

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestUser_DBError() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel")
	mockConfig.On("GetString", "auth.defaults.guard").Return("user")
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	app := NewApplication()
	token, err := app.LoginUsingID(1)
	assert.Nil(s.T(), err)

	mockCache := mock.Cache()
	mockCache.On("Get", "jwt:disabled:"+token, false).Return(false).Once()

	expired, err := app.Parse(token)
	assert.False(s.T(), expired)
	assert.Nil(s.T(), err)

	var user User

	mockOrm, mockDB, _ := mock.Orm()
	mockOrm.On("Query").Return(mockDB)
	mockDB.On("Find", &user, float64(1)).Return(errors.New("error")).Once()

	err = app.User(&user)
	assert.EqualError(s.T(), err, "error")

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestUser_Expired() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel")
	mockConfig.On("GetString", "auth.defaults.guard").Return("user")
	mockConfig.On("GetInt", "jwt.ttl").Return(2)

	app := NewApplication()
	token, err := app.LoginUsingID(1)
	assert.NotEmpty(s.T(), token)
	assert.Nil(s.T(), err)

	mockCache := mock.Cache()
	mockCache.On("Get", "jwt:disabled:"+token, false).Return(false).Once()

	time.Sleep(2 * unit)

	expired, err := app.Parse(token)
	assert.True(s.T(), expired)
	assert.Nil(s.T(), err)

	var user User
	err = app.User(&user)
	assert.EqualError(s.T(), err, "token expired")

	mockConfig.On("GetInt", "jwt.refresh_ttl").Return(2).Once()

	token, err = app.Refresh()
	assert.NotEmpty(s.T(), token)
	assert.Nil(s.T(), err)

	mockOrm, mockDB, _ := mock.Orm()
	mockOrm.On("Query").Return(mockDB)
	mockDB.On("Find", &user, float64(1)).Return(nil).Once()

	err = app.User(&user)
	assert.Nil(s.T(), err)

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestUser_RefreshExpired() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel")
	mockConfig.On("GetString", "auth.defaults.guard").Return("user")
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	app := NewApplication()
	token, err := app.LoginUsingID(1)
	assert.NotEmpty(s.T(), token)
	assert.Nil(s.T(), err)

	mockCache := mock.Cache()
	mockCache.On("Get", "jwt:disabled:"+token, false).Return(false).Once()

	time.Sleep(2 * unit)

	expired, err := app.Parse(token)
	assert.True(s.T(), expired)
	assert.Nil(s.T(), err)

	var user User
	err = app.User(&user)
	assert.EqualError(s.T(), err, "token expired")

	mockConfig.On("GetInt", "jwt.refresh_ttl").Return(1).Once()

	time.Sleep(2 * unit)

	token, err = app.Refresh()
	assert.Empty(s.T(), token)
	assert.EqualError(s.T(), err, "refresh time exceeded")

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestUser_Success() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel")
	mockConfig.On("GetString", "auth.defaults.guard").Return("user")
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	app := NewApplication()
	token, err := app.LoginUsingID(1)
	assert.Nil(s.T(), err)

	mockCache := mock.Cache()
	mockCache.On("Get", "jwt:disabled:"+token, false).Return(false).Once()

	expired, err := app.Parse(token)
	assert.False(s.T(), expired)
	assert.Nil(s.T(), err)

	var user User
	mockOrm, mockDB, _ := mock.Orm()
	mockOrm.On("Query").Return(mockDB)
	mockDB.On("Find", &user, float64(1)).Return(nil).Once()

	err = app.User(&user)
	assert.Nil(s.T(), err)

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestRefresh_NotParse() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "auth.defaults.guard").Return("user").Once()

	app := NewApplication()

	token, err := app.Refresh()
	assert.Empty(s.T(), token)
	assert.EqualError(s.T(), err, "parse token first")

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestRefresh_RefreshTimeExceeded() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel")
	mockConfig.On("GetString", "auth.defaults.guard").Return("user")
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	app := NewApplication()
	token, err := app.LoginUsingID(1)
	assert.Nil(s.T(), err)

	mockCache := mock.Cache()
	mockCache.On("Get", "jwt:disabled:"+token, false).Return(false).Once()

	expired, err := app.Parse(token)
	assert.False(s.T(), expired)
	assert.Nil(s.T(), err)

	mockConfig.On("GetInt", "jwt.refresh_ttl").Return(1).Once()
	time.Sleep(4 * unit)

	token, err = app.Refresh()
	assert.Empty(s.T(), token)
	assert.EqualError(s.T(), err, "refresh time exceeded")

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestRefresh_Success() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel")
	mockConfig.On("GetString", "auth.defaults.guard").Return("user")
	mockConfig.On("GetInt", "jwt.ttl").Return(2)

	app := NewApplication()
	token, err := app.LoginUsingID(1)
	assert.Nil(s.T(), err)

	mockCache := mock.Cache()
	mockCache.On("Get", "jwt:disabled:"+token, false).Return(false).Once()

	expired, err := app.Parse(token)
	assert.False(s.T(), expired)
	assert.Nil(s.T(), err)

	mockConfig.On("GetInt", "jwt.refresh_ttl").Return(1).Once()
	time.Sleep(2 * unit)

	token, err = app.Refresh()
	assert.NotEmpty(s.T(), token)
	assert.Nil(s.T(), err)

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLogout_CacheUnsupported() {
	app := NewApplication()
	assert.EqualError(s.T(), app.Logout(), "cache support is required")
}

func (s *AuthTestSuite) TestLogout_NotParse() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "auth.defaults.guard").Return("user").Once()

	app := NewApplication()
	_ = mock.Cache()
	assert.Nil(s.T(), app.Logout())

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLogout_SetDisabledCacheError() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel")
	mockConfig.On("GetString", "auth.defaults.guard").Return("user")
	mockConfig.On("GetInt", "jwt.ttl").Return(2)

	app := NewApplication()

	token, err := app.LoginUsingID(1)
	assert.Nil(s.T(), err)

	mockCache := mock.Cache()
	mockCache.On("Get", "jwt:disabled:"+token, false).Return(false).Once()

	expired, err := app.Parse(token)
	assert.False(s.T(), expired)
	assert.Nil(s.T(), err)

	mockCache.On("Put", testifymock.Anything, true, 2*unit).Return(errors.New("error")).Once()

	assert.EqualError(s.T(), app.Logout(), "error")

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLogout_Success() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel")
	mockConfig.On("GetString", "auth.defaults.guard").Return("user")
	mockConfig.On("GetInt", "jwt.ttl").Return(2)

	app := NewApplication()

	token, err := app.LoginUsingID(1)
	assert.NotEmpty(s.T(), token)
	assert.Nil(s.T(), err)

	mockCache := mock.Cache()
	mockCache.On("Get", "jwt:disabled:"+token, false).Return(false).Once()

	expired, err := app.Parse(token)
	assert.False(s.T(), expired)
	assert.Nil(s.T(), err)

	mockCache.On("Put", testifymock.Anything, true, 2*unit).Return(nil).Once()

	assert.Nil(s.T(), app.Logout())

	mockConfig.AssertExpectations(s.T())
}
