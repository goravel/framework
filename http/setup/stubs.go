package main

import (
	"strings"
)

type Stubs struct{}

func (s Stubs) CorsConfig(pkg, facadesImport, facadesPackage string) string {
	content := `package DummyPackage

import (
	"DummyFacadesImport"
)

func init() {
	config := DummyFacadesPackage.Config()
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

	content = strings.ReplaceAll(content, "DummyPackage", pkg)
	content = strings.ReplaceAll(content, "DummyFacadesImport", facadesImport)
	content = strings.ReplaceAll(content, "DummyFacadesPackage", facadesPackage)

	return content
}

func (s Stubs) HttpConfig(pkg, facadesImport, facadesPackage string) string {
	content := `package DummyPackage

import (
	"DummyFacadesImport"
)

func init() {
	config := DummyFacadesPackage.Config()
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
        
		// Default Client Name
		//
		// This determines which client is used when you call facades.Http() or
		// facades.Http().Client() without passing a specific name.
		"default_client": config.Env("HTTP_CLIENT_DEFAULT", "default"),
	
		// Client Configurations
		//
		// Here you may define multiple independent client configurations.
		// For example, you might have a "github" client with a specific base URL
		// and a "stripe" client with a longer timeout.
		"clients": map[string]any{
		   "default": map[string]any{
			  // The base URL for the client. All requests made using this client
			  // will be relative to this URL.
			  "base_url": config.Env("HTTP_CLIENT_BASE_URL", ""),
	
			  // The maximum amount of time a request can take, including connection
			  // establishment, redirects, and reading the response body.
			  "timeout": config.Env("HTTP_CLIENT_TIMEOUT", "30s"),
	
			  // The maximum number of idle (keep-alive) connections to keep across
			  // ALL hosts. Increasing this helps reuse TCP connections.
			  "max_idle_conns": config.Env("HTTP_CLIENT_MAX_IDLE_CONNS", 100),
	
			  // The maximum number of idle (keep-alive) connections to keep PER host.
			  // By default, Go sets this to 2, which is often a bottleneck.
			  // Increase this value for high-throughput applications.
			  "max_idle_conns_per_host": config.Env("HTTP_CLIENT_MAX_IDLE_CONNS_PER_HOST", 2),
	
			  // The maximum total number of connections (active + idle) allowed per host.
			  // A value of 0 means no limit.
			  "max_conns_per_host": config.Env("HTTP_CLIENT_MAX_CONN_PER_HOST", 0),
	
			  // The maximum amount of time an idle (keep-alive) connection will remain
			  // in the pool before closing itself.
			  "idle_conn_timeout": config.Env("HTTP_CLIENT_IDLE_CONN_TIMEOUT", "90s"),
		   },
		},
	})
}
`

	content = strings.ReplaceAll(content, "DummyPackage", pkg)
	content = strings.ReplaceAll(content, "DummyFacadesImport", facadesImport)
	content = strings.ReplaceAll(content, "DummyFacadesPackage", facadesPackage)

	return content
}

func (s Stubs) HttpFacade(pkg string) string {
	content := `package DummyPackage

import (
	"github.com/goravel/framework/contracts/http/client"
)

func Http() client.Request {
	return App().MakeHttp()
}
`

	return strings.ReplaceAll(content, "DummyPackage", pkg)
}

func (s Stubs) JwtConfig(pkg, facadesImport, facadesPackage string) string {
	content := `package DummyPackage

import (
	"DummyFacadesImport"
)

func init() {
	config := DummyFacadesPackage.Config()
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

	content = strings.ReplaceAll(content, "DummyPackage", pkg)
	content = strings.ReplaceAll(content, "DummyFacadesImport", facadesImport)
	content = strings.ReplaceAll(content, "DummyFacadesPackage", facadesPackage)

	return content
}

func (s Stubs) RateLimiterFacade(pkg string) string {
	content := `package DummyPackage

import (
	"github.com/goravel/framework/contracts/http"
)

func RateLimiter() http.RateLimiter {
	return App().MakeRateLimiter()
}
`

	return strings.ReplaceAll(content, "DummyPackage", pkg)
}

func (s Stubs) ViewFacade(pkg string) string {
	content := `package DummyPackage

import (
	"github.com/goravel/framework/contracts/http"
)

func View() http.View {
	return App().MakeView()
}
`

	return strings.ReplaceAll(content, "DummyPackage", pkg)
}
