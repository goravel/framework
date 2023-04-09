package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm/clause"

	contractsauth "github.com/goravel/framework/contracts/auth"
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/http"
	supporttime "github.com/goravel/framework/support/time"
	"github.com/goravel/framework/testing/mock"
)

var guard = "user"

type User struct {
	orm.Model
	Name string
}

type AuthTestSuite struct {
	suite.Suite
	auth *Auth
}

func TestAuthTestSuite(t *testing.T) {
	unit = time.Second
	suite.Run(t, &AuthTestSuite{
		auth: NewAuth(guard),
	})
}

func (s *AuthTestSuite) SetupTest() {

}

func (s *AuthTestSuite) TestLoginUsingID_EmptySecret() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("").Once()

	token, err := s.auth.LoginUsingID(http.Background(), 1)
	s.Empty(token)
	s.ErrorIs(err, ErrorEmptySecret)

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLoginUsingID_InvalidKey() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel").Once()
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	token, err := s.auth.LoginUsingID(http.Background(), "")
	s.Empty(token)
	s.ErrorIs(err, ErrorInvalidKey)

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLoginUsingID() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel").Once()
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	token, err := s.auth.LoginUsingID(http.Background(), 1)
	s.NotEmpty(token)
	s.Nil(err)

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLogin_Model() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel").Once()
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	var user User
	user.ID = 1
	user.Name = "Goravel"
	token, err := s.auth.Login(http.Background(), &user)
	s.NotEmpty(token)
	s.Nil(err)

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
	token, err := s.auth.Login(http.Background(), &user)
	s.NotEmpty(token)
	s.Nil(err)

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
	token, err := s.auth.Login(http.Background(), &errorUser)
	s.Empty(token)
	s.EqualError(err, "the primaryKey field was not found in the model, set primaryKey like orm.Model")
}

func (s *AuthTestSuite) TestLogin_NoPrimaryKey() {
	type User struct {
		ID   uint
		Name string
	}

	ctx := http.Background()
	var user User
	user.ID = 1
	user.Name = "Goravel"
	token, err := s.auth.Login(ctx, &user)
	s.Empty(token)
	s.ErrorIs(err, ErrorNoPrimaryKeyField)
}

func (s *AuthTestSuite) TestParse_TokenDisabled() {
	token := "1"
	mockCache, _, _ := mock.Cache()
	mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(true).Once()

	payload, err := s.auth.Parse(http.Background(), token)
	s.Nil(payload)
	s.EqualError(err, "token is disabled")
}

func (s *AuthTestSuite) TestParse_TokenInvalid() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel").Once()

	token := "1"
	mockCache, _, _ := mock.Cache()
	mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse(http.Background(), token)
	s.Nil(payload)
	s.NotNil(err)

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestParse_TokenExpired() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	now := supporttime.Now()
	ctx := http.Background()
	token, err := s.auth.LoginUsingID(ctx, 1)
	s.Nil(err)

	time.Sleep(2 * unit)

	mockCache, _, _ := mock.Cache()
	mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse(ctx, token)
	s.Equal(&contractsauth.Payload{
		Guard:    guard,
		Key:      "1",
		ExpireAt: jwt.NewNumericDate(now.Add(time.Duration(2) * unit)).Local(),
		IssuedAt: jwt.NewNumericDate(now).Local(),
	}, payload)
	s.ErrorIs(err, ErrorTokenExpired)

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestParse_InvalidCache() {
	facades.Cache = nil
	ctx := http.Background()
	payload, err := s.auth.Parse(ctx, "1")
	s.Nil(payload)
	s.EqualError(err, "cache support is required")
}

func (s *AuthTestSuite) TestParse_Success() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	ctx := http.Background()
	token, err := s.auth.LoginUsingID(ctx, 1)
	s.Nil(err)

	mockCache, _, _ := mock.Cache()
	mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse(ctx, token)
	s.Equal(&contractsauth.Payload{
		Guard:    guard,
		Key:      "1",
		ExpireAt: jwt.NewNumericDate(supporttime.Now().Add(time.Duration(2) * unit)).Local(),
		IssuedAt: jwt.NewNumericDate(supporttime.Now()).Local(),
	}, payload)
	s.Nil(err)

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestParse_SuccessWithPrefix() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	ctx := http.Background()
	token, err := s.auth.LoginUsingID(ctx, 1)
	s.Nil(err)

	mockCache, _, _ := mock.Cache()
	mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse(ctx, "Bearer "+token)
	s.Equal(&contractsauth.Payload{
		Guard:    guard,
		Key:      "1",
		ExpireAt: jwt.NewNumericDate(supporttime.Now().Add(time.Duration(2) * unit)).Local(),
		IssuedAt: jwt.NewNumericDate(supporttime.Now()).Local(),
	}, payload)
	s.Nil(err)

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestUser_NoParse() {
	mockConfig := mock.Config()

	ctx := http.Background()
	var user User
	err := s.auth.User(ctx, user)
	s.EqualError(err, "parse token first")

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestUser_DBError() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	ctx := http.Background()
	token, err := s.auth.LoginUsingID(ctx, 1)
	s.Nil(err)

	mockCache, _, _ := mock.Cache()
	mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse(ctx, token)
	s.NotNil(payload)
	s.Nil(err)

	var user User

	mockOrm, mockDB, _, _ := mock.Orm()
	mockOrm.On("Query").Return(mockDB)
	mockDB.On("Find", &user, clause.Eq{Column: clause.PrimaryColumn, Value: "1"}).Return(errors.New("error")).Once()

	err = s.auth.User(ctx, &user)
	s.EqualError(err, "error")

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestUser_Expired() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel").Times(3)
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Twice()

	ctx := http.Background()
	token, err := s.auth.LoginUsingID(ctx, 1)
	s.NotEmpty(token)
	s.Nil(err)

	mockCache, _, _ := mock.Cache()
	mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	time.Sleep(2 * unit)

	payload, err := s.auth.Parse(ctx, token)
	s.NotNil(payload)
	s.ErrorIs(err, ErrorTokenExpired)

	var user User
	err = s.auth.User(ctx, &user)
	s.EqualError(err, "token expired")

	mockConfig.On("GetInt", "jwt.refresh_ttl").Return(2).Once()

	token, err = s.auth.Refresh(ctx)
	s.NotEmpty(token)
	s.Nil(err)

	mockOrm, mockDB, _, _ := mock.Orm()
	mockOrm.On("Query").Return(mockDB)
	mockDB.On("Find", &user, clause.Eq{Column: clause.PrimaryColumn, Value: "1"}).Return(nil).Once()

	err = s.auth.User(ctx, &user)
	s.Nil(err)

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestUser_RefreshExpired() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	ctx := http.Background()
	token, err := s.auth.LoginUsingID(ctx, 1)
	s.NotEmpty(token)
	s.Nil(err)

	mockCache, _, _ := mock.Cache()
	mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	time.Sleep(2 * unit)

	payload, err := s.auth.Parse(ctx, token)
	s.NotNil(payload)
	s.ErrorIs(err, ErrorTokenExpired)

	var user User
	err = s.auth.User(ctx, &user)
	s.EqualError(err, "token expired")

	mockConfig.On("GetInt", "jwt.refresh_ttl").Return(1).Once()

	time.Sleep(2 * unit)

	token, err = s.auth.Refresh(ctx)
	s.Empty(token)
	s.EqualError(err, "refresh time exceeded")

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestUser_Success() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	ctx := http.Background()
	token, err := s.auth.LoginUsingID(ctx, 1)
	s.Nil(err)

	mockCache, _, _ := mock.Cache()
	mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse(ctx, token)
	s.NotNil(payload)
	s.Nil(err)

	var user User
	mockOrm, mockDB, _, _ := mock.Orm()
	mockOrm.On("Query").Return(mockDB)
	mockDB.On("Find", &user, clause.Eq{Column: clause.PrimaryColumn, Value: "1"}).Return(nil).Once()

	err = s.auth.User(ctx, &user)
	s.Nil(err)

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestRefresh_NotParse() {
	mockConfig := mock.Config()

	ctx := http.Background()
	token, err := s.auth.Refresh(ctx)
	s.Empty(token)
	s.EqualError(err, "parse token first")

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestRefresh_RefreshTimeExceeded() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	ctx := http.Background()
	token, err := s.auth.LoginUsingID(ctx, 2)

	s.Nil(err)

	mockCache, _, _ := mock.Cache()
	mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse(ctx, token)
	s.NotNil(payload)
	s.Nil(err)

	mockConfig.On("GetInt", "jwt.refresh_ttl").Return(1).Once()
	time.Sleep(4 * unit)

	token, err = s.auth.Refresh(ctx)
	s.Empty(token)
	s.EqualError(err, "refresh time exceeded")

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestRefresh_Success() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel").Times(3)
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Twice()

	ctx := http.Background()
	token, err := s.auth.LoginUsingID(ctx, 1)
	s.Nil(err)

	mockCache, _, _ := mock.Cache()
	mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse(ctx, token)
	s.NotNil(payload)
	s.Nil(err)

	mockConfig.On("GetInt", "jwt.refresh_ttl").Return(1).Once()
	time.Sleep(2 * unit)

	token, err = s.auth.Refresh(ctx)
	s.NotEmpty(token)
	s.Nil(err)

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLogout_CacheUnsupported() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel").Once()
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Once()

	ctx := http.Background()
	token, err := s.auth.LoginUsingID(ctx, 1)
	s.NotEmpty(token)
	s.Nil(err)
	s.EqualError(s.auth.Logout(ctx), "cache support is required")

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLogout_NotParse() {
	s.Nil(s.auth.Logout(http.Background()))
}

func (s *AuthTestSuite) TestLogout_SetDisabledCacheError() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Twice()

	ctx := http.Background()
	token, err := s.auth.LoginUsingID(ctx, 1)
	s.Nil(err)

	mockCache, _, _ := mock.Cache()
	mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse(ctx, token)
	s.NotNil(payload)
	s.Nil(err)

	mockCache.On("Put", testifymock.Anything, true, 2*unit).Return(errors.New("error")).Once()

	s.EqualError(s.auth.Logout(ctx), "error")

	mockConfig.AssertExpectations(s.T())
}

func (s *AuthTestSuite) TestLogout_Success() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "jwt.secret").Return("Goravel").Twice()
	mockConfig.On("GetInt", "jwt.ttl").Return(2).Twice()

	ctx := http.Background()
	token, err := s.auth.LoginUsingID(ctx, 1)
	s.NotEmpty(token)
	s.Nil(err)

	mockCache, _, _ := mock.Cache()
	mockCache.On("GetBool", "jwt:disabled:"+token, false).Return(false).Once()

	payload, err := s.auth.Parse(ctx, token)
	s.NotNil(payload)
	s.Nil(err)

	mockCache.On("Put", testifymock.Anything, true, 2*unit).Return(nil).Once()

	s.Nil(s.auth.Logout(ctx))

	mockConfig.AssertExpectations(s.T())
}
