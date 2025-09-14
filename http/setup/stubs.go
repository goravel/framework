package main

import (
	"strings"
)

type Stubs struct{}

func (s Stubs) CorsConfig(module string) string {
	content := `package config

import (
	"DummyModule/app/facades"
)

func init() {
	config := facades.Config()
	config.Add("cors", map[string]any{
		// Cross-Origin Resource Sharing (CORS) Configuration
		//
		// Here you may configure your settings for cross-origin resource sharing
		// or "CORS". This determines what cross-origin operations may execute
		// in web browsers. You are free to adjust these settings as needed.
		//
		// To learn more: https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS
		"paths":                []string{},
		"allowed_methods":      []string{"*"},
		"allowed_origins":      []string{"*"},
		"allowed_headers":      []string{"*"},
		"exposed_headers":      []string{},
		"max_age":              0,
		"supports_credentials": false,
	})
}
`

	return strings.ReplaceAll(content, "DummyModule", module)
}

func (s Stubs) HttpConfig(module string) string {
	content := `package config

import (
	"DummyModule/app/facades"
)

func init() {
	config := facades.Config()
	config.Add("http", map[string]any{
		// HTTP Driver
		"default": "",
		// HTTP Drivers
		"drivers": map[string]any{},
		// HTTP URL
		"url": config.Env("APP_URL", "http://localhost"),
		// HTTP Host
		"host": config.Env("APP_HOST", "127.0.0.1"),
		// HTTP Port
		"port": config.Env("APP_PORT", "3000"),
		// HTTP Timeout, default is 3 seconds
		"request_timeout": 3,
		// HTTPS Configuration
		"tls": map[string]any{
			// HTTPS Host
			"host": config.Env("APP_HOST", "127.0.0.1"),
			// HTTPS Port
			"port": config.Env("APP_PORT", "3000"),
			// SSL Certificate, you can put the certificate in /public folder
			"ssl": map[string]any{
				// ca.pem
				"cert": "",
				// ca.key
				"key": "",
			},
		},
		// HTTP Client Configuration
		"client": map[string]any{
			"base_url":                config.GetString("HTTP_CLIENT_BASE_URL"),
			"timeout":                 config.GetDuration("HTTP_CLIENT_TIMEOUT"),
			"max_idle_conns":          config.GetInt("HTTP_CLIENT_MAX_IDLE_CONNS"),
			"max_idle_conns_per_host": config.GetInt("HTTP_CLIENT_MAX_IDLE_CONNS_PER_HOST"),
			"max_conns_per_host":      config.GetInt("HTTP_CLIENT_MAX_CONN_PER_HOST"),
			"idle_conn_timeout":       config.GetDuration("HTTP_CLIENT_IDLE_CONN_TIMEOUT"),
		},
	})
}
`

	return strings.ReplaceAll(content, "DummyModule", module)
}

func (s Stubs) JwtConfig(module string) string {
	content := `package config

import (
	"DummyModule/app/facades"
)

func init() {
	config := facades.Config()
	config.Add("jwt", map[string]any{
		// JWT Authentication Secret
		//
		// Don't forget to set this in your .env file, as it will be used to sign
		// your tokens. A helper command is provided for this:
		// go run . artisan jwt:secret
		"secret": config.Env("JWT_SECRET", ""),

		// JWT time to live
		//
		// Specify the length of time (in minutes) that the token will be valid for.
		// Defaults to 1 hour.
		//
		// You can also set this to 0, to yield a never expiring token.
		// Some people may want this behaviour for e.g. a mobile app.
		// This is not particularly recommended, so make sure you have appropriate
		// systems in place to revoke the token if necessary.
		"ttl": config.Env("JWT_TTL", 60),

		// Refresh time to live
		//
		// Specify the length of time (in minutes) that the token can be refreshed
		// within. I.E. The user can refresh their token within a 2 week window of
		// the original token being created until they must re-authenticate.
		// Defaults to 2 weeks.
		//
		// You can also set this to 0, to yield an infinite refresh time.
		// Some may want this instead of never expiring tokens for e.g. a mobile app.
		// This is not particularly recommended, so make sure you have appropriate
		// systems in place to revoke the token if necessary.
		"refresh_ttl": config.Env("JWT_REFRESH_TTL", 20160),
	})
}
`

	return strings.ReplaceAll(content, "DummyModule", module)
}

func (s Stubs) HttpFacade() string {
	return `package facades

import (
	"github.com/goravel/framework/contracts/http/client"
)

func Http() client.Request {
	return App().MakeHttp()
}
`
}

func (s Stubs) RateLimiterFacade() string {
	return `package facades

import (
	"github.com/goravel/framework/contracts/http"
)

func RateLimiter() http.RateLimiter {
	return App().MakeRateLimiter()
}
`
}

func (s Stubs) ViewFacade() string {
	return `package facades

import (
	"github.com/goravel/framework/contracts/http"
)

func View() http.View {
	return App().MakeView()
}
`
}
