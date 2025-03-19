package auth

import (
	contractsauth "github.com/goravel/framework/contracts/auth"
	"github.com/goravel/framework/errors"
	mocksauth "github.com/goravel/framework/mocks/auth"
)

func (s *AuthTestSuite) TestCheck() {
	s.mockConfig.EXPECT().GetString("jwt.secret").Return("Goravel").Once()
	s.mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(0).Once()
	s.False(s.auth.Check())
	s.True(s.auth.Guest())
	token, err := s.auth.LoginUsingID(1)
	s.Nil(err)
	s.NotEmpty(token)
	s.True(s.auth.Check())
	s.False(s.auth.Guest())
}

func (s *AuthTestSuite) TestAuth_ExtendGuard() {
	s.auth.Extend("session", func(name string, a contractsauth.Auth, up contractsauth.UserProvider) (contractsauth.GuardDriver, error) {
		mockGuard := mocksauth.NewGuardDriver(s.T())
		mockGuard.EXPECT().ID().Return("session-id-xxxx", nil)
		return mockGuard, nil
	})

	s.mockConfig.EXPECT().GetString("auth.guards.admin.driver").Return("session").Once()
	s.mockConfig.EXPECT().GetString("auth.guards.admin.provider").Return("admin").Once()
	s.mockConfig.EXPECT().GetString("auth.providers.admin.driver").Return("orm").Once()

	guard, err := s.auth.Guard("admin")
	s.Nil(err)

	id, err := guard.ID()
	s.Nil(err)

	s.Equal("session-id-xxxx", id)
}

func (s *AuthTestSuite) TestAuth_ExtendProvider() {
	user := User{}
	mockProvider := mocksauth.NewUserProvider(s.T())
	mockProvider.EXPECT().RetriveByID(&user, "1").Return(nil).Run(func(user, id interface{}) {
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

	guard, err := s.auth.Guard("admin")
	s.Nil(err)

	authUser := User{}
	token, err := guard.LoginUsingID(1)
	s.Nil(err)
	s.NotEmpty(token)

	err = guard.User(&authUser)
	s.Nil(err)
	s.Equal("MockUser", authUser.Name)
	s.Equal(uint(1), authUser.ID)
}

func (s *AuthTestSuite) TestAuth_GuardDriverNotFoundException() {
	s.mockConfig.EXPECT().GetString("auth.guards.admin.driver").Return("unknown").Once()
	s.mockConfig.EXPECT().GetString("auth.guards.admin.provider").Return("admin").Once()
	s.mockConfig.EXPECT().GetString("auth.providers.admin.driver").Return("orm").Once()

	guard, err := s.auth.Guard("admin")
	s.Nil(guard)
	s.ErrorIs(err, errors.AuthGuardDriverNotFound)
}

func (s *AuthTestSuite) TestAuth_ProviderDriverNotFoundException() {
	s.mockConfig.EXPECT().GetString("auth.guards.admin.driver").Return("jwt").Once()
	s.mockConfig.EXPECT().GetString("auth.guards.admin.provider").Return("admin").Once()
	s.mockConfig.EXPECT().GetString("auth.providers.admin.driver").Return("unknown").Once()

	guard, err := s.auth.Guard("admin")
	s.Nil(guard)
	s.ErrorIs(err, errors.AuthProviderDriverNotFound)
}
