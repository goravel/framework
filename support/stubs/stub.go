package stubs

import "strings"

func DatabaseConfig(pkg, main string) string {
	content := `package DummyPackage

import (
	"DummyMain/app/facades"
)

func init() {
	config := facades.Config()
	config.Add("database", map[string]any{})
}
`

	return strings.ReplaceAll(strings.ReplaceAll(content, "DummyPackage", pkg), "DummyMain", main)
}
