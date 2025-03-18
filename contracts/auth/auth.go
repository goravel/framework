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
	LoginUsingID(id any) (token string, err error)
	// Refresh the token for the current user.
	Refresh() (token string, err error)
	// Logout logs the user out of the application.
	Logout() error
}

type Auth interface {
	Guard
	GetGuard(name string) (Guard, error)
	Extend(name string, fn GuardFunc)
	Provider(name string, fn UserProviderFunc)
}

type UserProvider interface {
	RetriveById(user any, id any) error
}

type UserProviderFunc func(auth Auth) (UserProvider, error)
type GuardFunc func(string, Auth, UserProvider) Guard

type Payload struct {
	Guard    string
	Key      string
	ExpireAt time.Time
	IssuedAt time.Time
}
