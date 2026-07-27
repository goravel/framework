package main

import (
	"strings"
)

type Stubs struct{}

func (s Stubs) BroadcastFacade(pkg string) string {
	content := `package DummyPackage

import (
	"github.com/goravel/framework/contracts/broadcasting"
)

func Broadcast() broadcasting.Broadcast {
	return App().MakeBroadcast()
}
`

	return strings.ReplaceAll(content, "DummyPackage", pkg)
}

func (s Stubs) Config(facadesImport, facadesPackage string) string {
	content := `package config

import (
	contractshttp "github.com/goravel/framework/contracts/http"
	"DummyFacadesImport"
)

func init() {
	config := DummyFacadesPackage.Config()
	config.Add("broadcasting", map[string]any{
		"default": config.Env("BROADCAST_CONNECTION", "log"),

		"connections": map[string]any{
			"pusher": map[string]any{
				"driver":  "pusher",
				"key":     config.Env("PUSHER_APP_KEY", ""),
				"secret":  config.Env("PUSHER_APP_SECRET", ""),
				"app_id":  config.Env("PUSHER_APP_ID", ""),
				"options": map[string]any{
					"cluster": config.Env("PUSHER_APP_CLUSTER", "mt1"),
					"host":    config.Env("PUSHER_HOST", ""),
					"port":    config.Env("PUSHER_PORT", 443),
					"scheme":  config.Env("PUSHER_SCHEME", "https"),
				},
			},
			"log": map[string]any{
				"driver": "log",
			},
			"null": map[string]any{
				"driver": "null",
			},
		},

		"auth": map[string]any{
			"enabled":    config.Env("BROADCAST_AUTH_ENABLED", true),
			"path":       config.Env("BROADCAST_AUTH_PATH", "/broadcasting/auth"),
			"middleware": []contractshttp.Middleware{},
		},
	})
}
`

	content = strings.ReplaceAll(content, "DummyFacadesImport", facadesImport)
	content = strings.ReplaceAll(content, "DummyFacadesPackage", facadesPackage)

	return content
}
