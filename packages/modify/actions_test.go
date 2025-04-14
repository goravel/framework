package modify

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	contractmatch "github.com/goravel/framework/contracts/packages/match"
	"github.com/goravel/framework/contracts/packages/modify"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/support/file"
)

type ModifyActionsTestSuite struct {
	suite.Suite
	content string
}

func (s *ModifyActionsTestSuite) SetupTest() {
	s.content = `package config

import (
	"github.com/goravel/framework/auth"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/crypt"
	"github.com/goravel/framework/facades"
)

func Boot() {}

func init() {
	config := facades.Config()
	config.Add("app", map[string]any{
		"name":  config.Env("APP_NAME", "Goravel"),
		"exist": map[string]any{},
		"providers": []foundation.ServiceProvider{
			&auth.AuthServiceProvider{},
			&crypt.ServiceProvider{},
		},
	})
}`
}

func (s *ModifyActionsTestSuite) TearDownTest() {}

func TestModifyActionsTestSuite(t *testing.T) {
	suite.Run(t, new(ModifyActionsTestSuite))
}

func (s *ModifyActionsTestSuite) TestActions() {
	tests := []struct {
		name     string
		matchers []contractmatch.GoNode
		actions  []modify.Action
		assert   func(filename string)
	}{
		{
			name:     "add config (not exist)",
			matchers: match.Config("app"),
			actions: []modify.Action{
				AddConfig("key", `"value"`),
			},
			assert: func(content string) {
				s.Contains(content, `func init() {
	config := facades.Config()
	config.Add("app", map[string]any{
		"name":  config.Env("APP_NAME", "Goravel"),
		"exist": map[string]any{},
		"providers": []foundation.ServiceProvider{
			&auth.AuthServiceProvider{},
			&crypt.ServiceProvider{},
		},
		"key": "value",
	})
}`)
			},
		},
		{
			name:     "add config (exist)",
			matchers: match.Config("app"),
			actions: []modify.Action{
				AddConfig("name", `"Goravel"`),
			},
			assert: func(content string) {
				s.NotContains(content, `"name": "Goravel"`)
			},
		},
		{
			name:     "add config (to map)",
			matchers: match.Config("app.exist"),
			actions: []modify.Action{
				AddConfig("key", `"value"`),
			},
			assert: func(content string) {
				s.Contains(content, `func init() {
	config := facades.Config()
	config.Add("app", map[string]any{
		"name": config.Env("APP_NAME", "Goravel"),
		"exist": map[string]any{
			"key": "value",
		},
		"providers": []foundation.ServiceProvider{
			&auth.AuthServiceProvider{},
			&crypt.ServiceProvider{},
		},
	})
}`)
			},
		},
		{
			name:     "add import",
			matchers: []contractmatch.GoNode{match.Imports()},
			actions: []modify.Action{
				AddImport("github.com/goravel/test", "t"),
			},
			assert: func(content string) {
				s.Contains(content, `import (
	"github.com/goravel/framework/auth"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/crypt"
	"github.com/goravel/framework/facades"
	t "github.com/goravel/test"
)`)
			},
		},
		{
			name:     "add provider (not exist)",
			matchers: match.Providers(),
			actions: []modify.Action{
				AddProvider("&test.ServiceProvider{}"),
			},
			assert: func(content string) {
				s.Contains(content, `func init() {
	config := facades.Config()
	config.Add("app", map[string]any{
		"name":  config.Env("APP_NAME", "Goravel"),
		"exist": map[string]any{},
		"providers": []foundation.ServiceProvider{
			&auth.AuthServiceProvider{},
			&crypt.ServiceProvider{},
			&test.ServiceProvider{},
		},
	})
}`)
			},
		},
		{
			name:     "add provider (exist)",
			matchers: match.Providers(),
			actions: []modify.Action{
				AddProvider("&crypt.ServiceProvider{}"),
			},
			assert: func(content string) {
				s.Contains(content, `func init() {
	config := facades.Config()
	config.Add("app", map[string]any{
		"name":  config.Env("APP_NAME", "Goravel"),
		"exist": map[string]any{},
		"providers": []foundation.ServiceProvider{
			&auth.AuthServiceProvider{},
			&crypt.ServiceProvider{},
		},
	})
}`)
			},
		},
		{
			name:     "add provider before",
			matchers: match.Providers(),
			actions: []modify.Action{
				AddProvider("&test.ServiceProvider{}", "&auth.AuthServiceProvider{}"),
			},
			assert: func(content string) {
				s.Contains(content, `func init() {
	config := facades.Config()
	config.Add("app", map[string]any{
		"name":  config.Env("APP_NAME", "Goravel"),
		"exist": map[string]any{},
		"providers": []foundation.ServiceProvider{
			&test.ServiceProvider{},
			&auth.AuthServiceProvider{},
			&crypt.ServiceProvider{},
		},
	})
}`)
			},
		},
		{
			name:     "remove config",
			matchers: match.Config("app"),
			actions: []modify.Action{
				RemoveConfig("providers"),
			},
			assert: func(content string) {
				s.NotContains(content, "providers")
			},
		},
		{
			name:     "remove import",
			matchers: []contractmatch.GoNode{match.Imports()},
			actions: []modify.Action{
				RemoveImport("github.com/goravel/framework/auth"),
			},
			assert: func(content string) {
				s.NotContains(content, `"github.com/goravel/framework/auth"`)
			},
		},
		{
			name:     "remove provider",
			matchers: match.Providers(),
			actions: []modify.Action{
				RemoveProvider("&auth.AuthServiceProvider{}"),
			},
			assert: func(content string) {
				s.NotContains(content, "&auth.AuthServiceProvider{}")
			},
		},
		{
			name:     "replace config",
			matchers: match.Config("app"),
			actions: []modify.Action{
				ReplaceConfig("name", `"Goravel"`),
			},
			assert: func(content string) {
				s.Contains(content, `"name":  "Goravel"`)
				s.NotContains(content, `config.Env("APP_NAME", "Goravel")`)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			sourceFile := filepath.Join(s.T().TempDir(), "test.go")
			s.Require().NoError(file.PutContent(sourceFile, s.content))
			s.Require().NoError(File(sourceFile).Find(tt.matchers...).Modify(tt.actions...).Apply())
			content, err := file.GetContent(sourceFile)
			s.Require().NoError(err)
			tt.assert(content)
		})
	}
}
