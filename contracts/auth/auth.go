package auth

import (
	"time"
)

type Guard interface {
	// Check whether user logged in or not
	Check() bool

	// Check whether user *not* logged in or not | !Check()
	Guest() bool

	// User returns the current authenticated user.
	User(user any) error
	// ID returns the current user id.
	ID() (string, error)
	// Login logs a user into the application.
	Login(user any) (err error)
	// LoginUsingID logs the given user ID into the application.
	LoginUsingID(id any) (err error)
	// Logout logs the user out of the application.
	Logout() error
}

type Auth interface {
	Guard
	// Guard attempts to get the guard against the local cache.
	GetGuard(name string) (Guard, error)
}

type UserProvider interface {
	RetriveById(any) (any, error)
	RetriveByCredentials(map[string]any) (any, error)
}

type UserProviderFunc func(Auth) UserProvider
type AuthGuardFunc func(string, Auth, UserProvider) Guard

type Payload struct {
	Guard    string
	Key      string
	ExpireAt time.Time
	IssuedAt time.Time
}
