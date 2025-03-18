package packages

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/packages"
	"github.com/goravel/framework/support/file"
)

type HelperTestSuite struct {
	suite.Suite
	content string
}

func (s *HelperTestSuite) SetupTest() {
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

func (s *HelperTestSuite) TearDownTest() {}

func TestHelperTestSuite(t *testing.T) {
	suite.Run(t, new(HelperTestSuite))
}

func (s *HelperTestSuite) TestHelper() {
	cases := []struct {
		name      string
		modifiers []packages.GoNodeModifier

		assert func(filename string)
	}{
		{
			name: "AddConfigSpec(not exist)",
			modifiers: []packages.GoNodeModifier{
				AddConfigSpec("app", "key", `"value"`),
			},
			assert: func(content string) {
				s.Contains(content, `"key": "value",`)
			},
		},
		{
			name: "AddConfigSpec(exist)",
			modifiers: []packages.GoNodeModifier{
				AddConfigSpec("app", "name", `"Goravel"`),
			},
			assert: func(content string) {
				s.NotContains(content, `"name": "Goravel"`)
			},
		},
		{
			name: "AddConfigSpec(to map)",
			modifiers: []packages.GoNodeModifier{
				AddConfigSpec("app.exist", "key", `"value"`),
			},
			assert: func(content string) {
				s.Contains(content, `"key": "value"`)
			},
		},
		{
			name: "AddImportSpec",
			modifiers: []packages.GoNodeModifier{
				AddImportSpec("github.com/goravel/test", "t"),
			},
			assert: func(content string) {
				s.Contains(content, `t "github.com/goravel/test"`)
			},
		},
		{
			name: "AddProviderSpec(not exist)",
			modifiers: []packages.GoNodeModifier{
				AddProviderSpec("&test.ServiceProvider{}"),
			},
			assert: func(content string) {
				s.Contains(content, "&crypt.ServiceProvider{},\n\t\t\t&test.ServiceProvider{}")
			},
		},
		{
			name: "AddProviderSpec(exist)",
			modifiers: []packages.GoNodeModifier{
				AddProviderSpec("&crypt.ServiceProvider{}"),
			},
			assert: func(content string) {
				s.Contains(content, "&crypt.ServiceProvider{}")
			},
		},
		{
			name: "AddProviderSpecAfter",
			modifiers: []packages.GoNodeModifier{
				AddProviderSpecAfter("&test.ServiceProvider{}", "&auth.AuthServiceProvider{}"),
			},
			assert: func(content string) {
				s.Contains(content, "&auth.AuthServiceProvider{},\n\t\t\t&test.ServiceProvider{}")
			},
		},
		{
			name: "AddProviderSpecBefore",
			modifiers: []packages.GoNodeModifier{
				AddProviderSpecBefore("&test.ServiceProvider{}", "&auth.AuthServiceProvider{}"),
			},
			assert: func(content string) {
				s.Contains(content, "&test.ServiceProvider{},\n\t\t\t&auth.AuthServiceProvider{}")
			},
		},
		{
			name: "RemoveConfigSpec",
			modifiers: []packages.GoNodeModifier{
				RemoveConfigSpec("app.providers"),
			},
			assert: func(content string) {
				s.NotContains(content, "providers")
			},
		},
		{
			name: "RemoveImportSpec",
			modifiers: []packages.GoNodeModifier{
				RemoveImportSpec("github.com/goravel/framework/auth"),
			},
			assert: func(content string) {
				s.NotContains(content, `"github.com/goravel/framework/auth"`)
			},
		},
		{
			name: "RemoveProviderSpec",
			modifiers: []packages.GoNodeModifier{
				RemoveProviderSpec("&auth.AuthServiceProvider{}"),
			},
			assert: func(content string) {
				s.NotContains(content, "&auth.AuthServiceProvider{}")
			},
		},
		{
			name: "ReplaceConfigSpec",
			modifiers: []packages.GoNodeModifier{
				ReplaceConfigSpec("app.name", `"Goravel"`),
			},
			assert: func(content string) {
				s.Contains(content, `"name":  "Goravel"`)
				s.NotContains(content, `config.Env("APP_NAME", "Goravel")`)
			},
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			dir := s.T().TempDir()
			filename := "test.go"
			s.Require().NoError(file.PutContent(filepath.Join(dir, filename), s.content))
			mg := ModifyGoFile{
				File:      filename,
				Modifiers: tc.modifiers,
			}
			s.Require().NoError(mg.Apply(dir))
			content, err := file.GetContent(filepath.Join(dir, filename))
			s.Require().NoError(err)
			tc.assert(content)
		})
	}
}
