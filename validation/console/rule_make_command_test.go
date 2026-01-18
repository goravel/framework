package console

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
)

var ruleValidationServiceProvider = `package providers

import (
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/facades"
)

type ValidationServiceProvider struct {
}

func (receiver *ValidationServiceProvider) Register(app foundation.Application) {

}

func (receiver *ValidationServiceProvider) Boot(app foundation.Application) {
	if err := facades.Validation().AddRules(receiver.rules()); err != nil {
		facades.Log().Errorf("add rules error: %+v", err)
	}
}

func (receiver *ValidationServiceProvider) rules() []validation.Rule {
	return []validation.Rule{}
}
`

var bootstrapApp = `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`

var bootstrapAppWithRules = `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithRules(Rules).WithConfig(config.Boot).Start()
}
`

var bootstrapRules = `package bootstrap

import (
	"github.com/goravel/framework/contracts/validation"

	"goravel/app/rules"
)

func Rules() []validation.Rule {
	return []validation.Rule{
		&rules.ExistingRule{},
	}
}
`

func TestRuleMakeCommand(t *testing.T) {
	ruleMakeCommand := &RuleMakeCommand{}
	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Argument(0).Return("").Once()
	mockContext.EXPECT().Ask("Enter the rule name", mock.Anything).Return("", errors.New("the rule name cannot be empty")).Once()
	mockContext.EXPECT().Error("the rule name cannot be empty").Once()
	assert.NoError(t, ruleMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("Uppercase").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Rule created successfully").Once()
	mockContext.EXPECT().Error(mock.MatchedBy(func(msg string) bool {
		return strings.Contains(msg, "rule register failed:")
	})).Once()
	assert.NoError(t, ruleMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/rules/uppercase.go"))

	mockContext.On("Argument", 0).Return("Uppercase").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	mockContext.EXPECT().Error("the rule already exists. Use the --force or -f flag to overwrite").Once()
	assert.NoError(t, ruleMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("User/Phone").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Rule created successfully").Once()
	mockContext.EXPECT().Success("Rule registered successfully").Once()
	assert.NoError(t, file.PutContent("app/providers/validation_service_provider.go", ruleValidationServiceProvider))
	assert.NoError(t, ruleMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/rules/User/phone.go"))
	assert.True(t, file.Contain("app/rules/User/phone.go", "package User"))
	assert.True(t, file.Contain("app/rules/User/phone.go", "type Phone struct"))
	assert.True(t, file.Contain("app/rules/User/phone.go", "user_phone"))
	assert.True(t, file.Contain("app/providers/validation_service_provider.go", "app/rules/User"))
	assert.True(t, file.Contain("app/providers/validation_service_provider.go", "&User.Phone{}"))
	assert.Nil(t, file.Remove("app"))
}

func TestRuleMakeCommand_WithBootstrapSetup(t *testing.T) {
	tests := []struct {
		name                string
		ruleName            string
		bootstrapAppSetup   string
		bootstrapRulesSetup string
		expectRulesFile     bool
		expectSuccess       bool
	}{
		{
			name:              "creates rule and registers in bootstrap without existing rules.go",
			ruleName:          "Uppercase",
			bootstrapAppSetup: bootstrapApp,
			expectRulesFile:   true,
			expectSuccess:     true,
		},
		{
			name:                "creates rule and appends to existing rules.go",
			ruleName:            "Lowercase",
			bootstrapAppSetup:   bootstrapAppWithRules,
			bootstrapRulesSetup: bootstrapRules,
			expectRulesFile:     true,
			expectSuccess:       true,
		},
		{
			name:              "creates nested rule in bootstrap setup",
			ruleName:          "User/Phone",
			bootstrapAppSetup: bootstrapApp,
			expectRulesFile:   true,
			expectSuccess:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup bootstrap directory
			bootstrapDir := support.Config.Paths.Bootstrap
			appFile := filepath.Join(bootstrapDir, "app.go")
			rulesFile := filepath.Join(bootstrapDir, "rules.go")

			// Create bootstrap/app.go
			assert.NoError(t, file.PutContent(appFile, tt.bootstrapAppSetup))
			defer func() {
				assert.NoError(t, file.Remove(bootstrapDir))
				assert.NoError(t, file.Remove("app"))
			}()

			// Create bootstrap/rules.go if provided
			if tt.bootstrapRulesSetup != "" {
				assert.NoError(t, file.PutContent(rulesFile, tt.bootstrapRulesSetup))
			}

			// Setup mock context
			ruleMakeCommand := &RuleMakeCommand{}
			mockContext := mocksconsole.NewContext(t)
			mockContext.EXPECT().Argument(0).Return(tt.ruleName).Once()
			mockContext.EXPECT().OptionBool("force").Return(false).Once()
			mockContext.EXPECT().Success("Rule created successfully").Once()

			if tt.expectSuccess {
				mockContext.EXPECT().Success("Rule registered successfully").Once()
			} else {
				mockContext.EXPECT().Error(mock.MatchedBy(func(msg string) bool {
					return strings.Contains(msg, "rule register failed:")
				})).Once()
			}

			// Execute command
			assert.NoError(t, ruleMakeCommand.Handle(mockContext))

			// Verify rule file was created using the same logic as Make.GetFilePath()
			var rulePath string
			if strings.Contains(tt.ruleName, "/") {
				parts := strings.Split(tt.ruleName, "/")
				folderPath := filepath.Join(parts[:len(parts)-1]...)
				structName := str.Of(parts[len(parts)-1]).Studly().String()
				fileName := str.Of(structName).Snake().String() + ".go"
				rulePath = filepath.Join("app/rules", folderPath, fileName)
			} else {
				structName := str.Of(tt.ruleName).Studly().String()
				fileName := str.Of(structName).Snake().String() + ".go"
				rulePath = filepath.Join("app/rules", fileName)
			}
			assert.True(t, file.Exists(rulePath), "Rule file should exist at %s", rulePath)

			// Verify bootstrap/rules.go was created/updated
			if tt.expectRulesFile {
				assert.True(t, file.Exists(rulesFile), "bootstrap/rules.go should exist")
				rulesContent, err := file.GetContent(rulesFile)
				assert.NoError(t, err)
				assert.Contains(t, rulesContent, "func Rules() []validation.Rule")
			}

			// Verify bootstrap/app.go contains WithRules
			appContent, err := file.GetContent(appFile)
			assert.NoError(t, err)
			if tt.expectSuccess {
				assert.Contains(t, appContent, "WithRules(Rules)")
			}
		})
	}
}

func TestRuleMakeCommand_WithNonBootstrapSetup(t *testing.T) {
	tests := []struct {
		name          string
		ruleName      string
		expectSuccess bool
	}{
		{
			name:          "creates rule and registers in validation service provider",
			ruleName:      "Email",
			expectSuccess: true,
		},
		{
			name:          "creates nested rule in validation service provider",
			ruleName:      "Custom/Alphanumeric",
			expectSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup validation service provider without bootstrap setup
			providerPath := filepath.Join("app", "providers", "validation_service_provider.go")
			assert.NoError(t, file.PutContent(providerPath, ruleValidationServiceProvider))
			defer func() {
				assert.NoError(t, file.Remove("app"))
			}()

			// Setup mock context
			ruleMakeCommand := &RuleMakeCommand{}
			mockContext := mocksconsole.NewContext(t)
			mockContext.EXPECT().Argument(0).Return(tt.ruleName).Once()
			mockContext.EXPECT().OptionBool("force").Return(false).Once()
			mockContext.EXPECT().Success("Rule created successfully").Once()

			if tt.expectSuccess {
				mockContext.EXPECT().Success("Rule registered successfully").Once()
			} else {
				mockContext.EXPECT().Error(mock.MatchedBy(func(msg string) bool {
					return strings.Contains(msg, "rule register failed:")
				})).Once()
			}

			// Execute command
			assert.NoError(t, ruleMakeCommand.Handle(mockContext))

			// Verify rule file was created using the same logic as Make.GetFilePath()
			var rulePath string
			if strings.Contains(tt.ruleName, "/") {
				parts := strings.Split(tt.ruleName, "/")
				folderPath := filepath.Join(parts[:len(parts)-1]...)
				structName := str.Of(parts[len(parts)-1]).Studly().String()
				fileName := str.Of(structName).Snake().String() + ".go"
				rulePath = filepath.Join("app/rules", folderPath, fileName)
			} else {
				structName := str.Of(tt.ruleName).Studly().String()
				fileName := str.Of(structName).Snake().String() + ".go"
				rulePath = filepath.Join("app/rules", fileName)
			}
			assert.True(t, file.Exists(rulePath))

			// Verify registration in validation service provider
			if tt.expectSuccess {
				providerContent, err := file.GetContent(providerPath)
				assert.NoError(t, err)
				assert.Contains(t, providerContent, "app/rules")
			}
		})
	}
}

func TestRuleMakeCommand_RegistrationError(t *testing.T) {
	t.Run("handles error when bootstrap rules.go exists but WithRules not registered", func(t *testing.T) {
		// Setup bootstrap directory with app.go but no WithRules
		bootstrapDir := support.Config.Paths.Bootstrap
		appFile := filepath.Join(bootstrapDir, "app.go")
		rulesFile := filepath.Join(bootstrapDir, "rules.go")

		// Create bootstrap/app.go without WithRules
		assert.NoError(t, file.PutContent(appFile, bootstrapApp))
		// Create rules.go that shouldn't exist without WithRules
		assert.NoError(t, file.PutContent(rulesFile, bootstrapRules))
		defer func() {
			assert.NoError(t, file.Remove(bootstrapDir))
			assert.NoError(t, file.Remove("app"))
		}()

		// Setup mock context
		ruleMakeCommand := &RuleMakeCommand{}
		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Argument(0).Return("FailRule").Once()
		mockContext.EXPECT().OptionBool("force").Return(false).Once()
		mockContext.EXPECT().Success("Rule created successfully").Once()
		mockContext.EXPECT().Error(mock.MatchedBy(func(msg string) bool {
			return strings.Contains(msg, "rule register failed:")
		})).Once()

		// Execute command
		assert.NoError(t, ruleMakeCommand.Handle(mockContext))
	})

	t.Run("handles error when validation service provider is malformed", func(t *testing.T) {
		// Setup malformed validation service provider
		providerPath := filepath.Join("app", "providers", "validation_service_provider.go")
		malformedProvider := `package providers
// This is a malformed file without proper structure
`
		assert.NoError(t, file.PutContent(providerPath, malformedProvider))
		defer func() {
			assert.NoError(t, file.Remove("app"))
		}()

		// Setup mock context
		ruleMakeCommand := &RuleMakeCommand{}
		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Argument(0).Return("ErrorRule").Once()
		mockContext.EXPECT().OptionBool("force").Return(false).Once()
		mockContext.EXPECT().Success("Rule created successfully").Once()
		mockContext.EXPECT().Error(mock.MatchedBy(func(msg string) bool {
			return strings.Contains(msg, "rule register failed:")
		})).Once()

		// Execute command
		assert.NoError(t, ruleMakeCommand.Handle(mockContext))
	})
}
