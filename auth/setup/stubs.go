package main

import (
	"strings"
)

type Stubs struct{}

func (s Stubs) Config(module string) string {
	content := `package config

import (
	"DummyModule/app/facades"
)

func init() {
	config := facades.Config()
	config.Add("auth", map[string]any{
		// Authentication Defaults
		//
		// This option controls the default authentication "guard"
		// reset options for your application. You may change these defaults
		// as required, but they're a perfect start for most applications.
		"defaults": map[string]any{
			"guard": "user",
		},

		// Authentication Guards
		//
		// Next, you may define every authentication guard for your application.
		// Of course, a great default configuration has been defined for you
		// here which uses session storage and the Eloquent user provider.
		//
		// All authentication drivers have a user provider. This defines how the
		// users are actually retrieved out of your database or other storage
		// mechanisms used by this application to persist your user's data.
		//
		// Supported drivers: "jwt", "session"
		"guards": map[string]any{
			"user": map[string]any{
				"driver":   "jwt",
				"provider": "user",
			},
		},

		// Supported: "orm"
		"providers": map[string]any{
			"user": map[string]any{
				"driver": "orm",
			},
		},
	})
}
`

	return strings.ReplaceAll(content, "DummyModule", module)
}

func (s Stubs) AuthFacade() string {
	return `package facades

import (
	"github.com/goravel/framework/contracts/auth"
	"github.com/goravel/framework/contracts/http"
)

func Auth(ctx ...http.Context) auth.Auth {
	return App().MakeAuth(ctx...)
}
`
}

func (s Stubs) GateFacade() string {
	return `package facades

import (
	"github.com/goravel/framework/contracts/auth/access"
	"github.com/goravel/framework/contracts/http"
)

func Gate(ctx ...http.Context) access.Gate {
	return App().MakeGate()
}
`
}
