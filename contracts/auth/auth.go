package auth

import (
	"time"
)

type GuardDriver interface {
	// Check whether user logged in or not
	Check() bool
	// Check whether user *not* logged in or not | !Check()
	Guest() bool
	// User returns the current authenticated user.
	User(user any) error
	// ID returns the current user id.
	ID() (token string, err error)
	// Login logs a user into the application.
	Login(user any) (token string, err error)
	// LoginUsingID logs the given user ID into the application.
	LoginUsingID(id any) (token string, err error)
	// Parse the given token.
	Parse(token string) (*Payload, error)
	// Refresh the token for the current user.
	Refresh() (token string, err error)
	// Logout logs the user out of the application.
	Logout() error
}

type Auth interface {
	GuardDriver
	Guard(name string) GuardDriver
	Extend(name string, fn GuardFunc)
	Provider(name string, fn UserProviderFunc)
}

type UserProvider interface {
	RetriveByID(user any, id any) error
	GetID(user any) any
}

type UserProviderFunc func(auth Auth) (UserProvider, error)
type GuardFunc func(name string, auth Auth, userProvider UserProvider) (guard GuardDriver, err error)

type Payload struct {
	Guard    string
	Key      string
	ExpireAt time.Time
	IssuedAt time.Time
}
