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

var filterValidationServiceProvider = `package providers

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
	if err := facades.Validation().AddFilters(receiver.filters()); err != nil {
		facades.Log().Errorf("add filters error: %+v", err)
	}
}

func (receiver *ValidationServiceProvider) rules() []validation.Rule {
	return []validation.Rule{}
}

func (receiver *ValidationServiceProvider) filters() []validation.Filter {
	return []validation.Filter{}
}
`

var bootstrapAppFilter = `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`

var bootstrapAppWithFilters = `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithFilters(Filters).WithConfig(config.Boot).Run()
}
`

var bootstrapFilters = `package bootstrap

import (
	"github.com/goravel/framework/contracts/validation"

	"goravel/app/filters"
)

func Filters() []validation.Filter {
	return []validation.Filter{
		&filters.ExistingFilter{},
	}
}
`

func TestFilterMakeCommand(t *testing.T) {
	filterMakeCommand := &FilterMakeCommand{}
	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Argument(0).Return("").Once()
	mockContext.EXPECT().Ask("Enter the filter name", mock.Anything).Return("", errors.New("the filter name cannot be empty")).Once()
	mockContext.EXPECT().Error("the filter name cannot be empty").Once()
	assert.NoError(t, filterMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("Uppercase").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Filter created successfully").Once()
	mockContext.EXPECT().Error(mock.MatchedBy(func(msg string) bool {
		return strings.Contains(msg, "filter register failed:")
	})).Once()
	assert.NoError(t, filterMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/filters/uppercase.go"))

	mockContext.On("Argument", 0).Return("Uppercase").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	mockContext.EXPECT().Error("the filter already exists. Use the --force or -f flag to overwrite").Once()
	assert.NoError(t, filterMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("User/Phone").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Filter created successfully").Once()
	mockContext.EXPECT().Success("Filter registered successfully").Once()
	assert.NoError(t, file.PutContent("app/providers/validation_service_provider.go", filterValidationServiceProvider))
	assert.NoError(t, filterMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/filters/User/phone.go"))
	assert.True(t, file.Contain("app/filters/User/phone.go", "package User"))
	assert.True(t, file.Contain("app/filters/User/phone.go", "type Phone struct"))
	assert.True(t, file.Contain("app/filters/User/phone.go", "user_phone"))
	assert.True(t, file.Contain("app/providers/validation_service_provider.go", "app/filters/User"))
	assert.True(t, file.Contain("app/providers/validation_service_provider.go", "&User.Phone{}"))
	assert.Nil(t, file.Remove("app"))
}

func TestFilterMakeCommand_WithBootstrapSetup(t *testing.T) {
	tests := []struct {
		name                  string
		filterName            string
		bootstrapAppSetup     string
		bootstrapFiltersSetup string
		expectFiltersFile     bool
		expectSuccess         bool
	}{
		{
			name:              "creates filter and registers in bootstrap without existing filters.go",
			filterName:        "Uppercase",
			bootstrapAppSetup: bootstrapAppFilter,
			expectFiltersFile: true,
			expectSuccess:     true,
		},
		{
			name:                  "creates filter and appends to existing filters.go",
			filterName:            "Lowercase",
			bootstrapAppSetup:     bootstrapAppWithFilters,
			bootstrapFiltersSetup: bootstrapFilters,
			expectFiltersFile:     true,
			expectSuccess:         true,
		},
		{
			name:              "creates nested filter in bootstrap setup",
			filterName:        "User/Phone",
			bootstrapAppSetup: bootstrapAppFilter,
			expectFiltersFile: true,
			expectSuccess:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup bootstrap directory
			bootstrapDir := support.Config.Paths.Bootstrap
			appFile := filepath.Join(bootstrapDir, "app.go")
			filtersFile := filepath.Join(bootstrapDir, "filters.go")

			// Create bootstrap/app.go
			assert.NoError(t, file.PutContent(appFile, tt.bootstrapAppSetup))
			defer func() {
				assert.NoError(t, file.Remove(bootstrapDir))
				assert.NoError(t, file.Remove("app"))
			}()

			// Create bootstrap/filters.go if provided
			if tt.bootstrapFiltersSetup != "" {
				assert.NoError(t, file.PutContent(filtersFile, tt.bootstrapFiltersSetup))
			}

			// Setup mock context
			filterMakeCommand := &FilterMakeCommand{}
			mockContext := mocksconsole.NewContext(t)
			mockContext.EXPECT().Argument(0).Return(tt.filterName).Once()
			mockContext.EXPECT().OptionBool("force").Return(false).Once()
			mockContext.EXPECT().Success("Filter created successfully").Once()

			if tt.expectSuccess {
				mockContext.EXPECT().Success("Filter registered successfully").Once()
			} else {
				mockContext.EXPECT().Error(mock.MatchedBy(func(msg string) bool {
					return strings.Contains(msg, "filter register failed:")
				})).Once()
			}

			// Execute command
			assert.NoError(t, filterMakeCommand.Handle(mockContext))

			// Verify filter file was created using the same logic as Make.GetFilePath()
			var filterPath string
			if strings.Contains(tt.filterName, "/") {
				parts := strings.Split(tt.filterName, "/")
				folderPath := filepath.Join(parts[:len(parts)-1]...)
				structName := str.Of(parts[len(parts)-1]).Studly().String()
				fileName := str.Of(structName).Snake().String() + ".go"
				filterPath = filepath.Join("app/filters", folderPath, fileName)
			} else {
				structName := str.Of(tt.filterName).Studly().String()
				fileName := str.Of(structName).Snake().String() + ".go"
				filterPath = filepath.Join("app/filters", fileName)
			}
			assert.True(t, file.Exists(filterPath), "Filter file should exist at %s", filterPath)

			// Verify bootstrap/filters.go was created/updated
			if tt.expectFiltersFile {
				assert.True(t, file.Exists(filtersFile), "bootstrap/filters.go should exist")
				filtersContent, err := file.GetContent(filtersFile)
				assert.NoError(t, err)
				assert.Contains(t, filtersContent, "func Filters() []validation.Filter")
			}

			// Verify bootstrap/app.go contains WithFilters
			appContent, err := file.GetContent(appFile)
			assert.NoError(t, err)
			if tt.expectSuccess {
				assert.Contains(t, appContent, "WithFilters(Filters)")
			}
		})
	}
}

func TestFilterMakeCommand_WithNonBootstrapSetup(t *testing.T) {
	tests := []struct {
		name          string
		filterName    string
		expectSuccess bool
	}{
		{
			name:          "creates filter and registers in validation service provider",
			filterName:    "Email",
			expectSuccess: true,
		},
		{
			name:          "creates nested filter in validation service provider",
			filterName:    "Custom/Alphanumeric",
			expectSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup validation service provider without bootstrap setup
			providerPath := filepath.Join("app", "providers", "validation_service_provider.go")
			assert.NoError(t, file.PutContent(providerPath, filterValidationServiceProvider))
			defer func() {
				assert.NoError(t, file.Remove("app"))
			}()

			// Setup mock context
			filterMakeCommand := &FilterMakeCommand{}
			mockContext := mocksconsole.NewContext(t)
			mockContext.EXPECT().Argument(0).Return(tt.filterName).Once()
			mockContext.EXPECT().OptionBool("force").Return(false).Once()
			mockContext.EXPECT().Success("Filter created successfully").Once()

			if tt.expectSuccess {
				mockContext.EXPECT().Success("Filter registered successfully").Once()
			} else {
				mockContext.EXPECT().Error(mock.MatchedBy(func(msg string) bool {
					return strings.Contains(msg, "filter register failed:")
				})).Once()
			}

			// Execute command
			assert.NoError(t, filterMakeCommand.Handle(mockContext))

			// Verify filter file was created using the same logic as Make.GetFilePath()
			var filterPath string
			if strings.Contains(tt.filterName, "/") {
				parts := strings.Split(tt.filterName, "/")
				folderPath := filepath.Join(parts[:len(parts)-1]...)
				structName := str.Of(parts[len(parts)-1]).Studly().String()
				fileName := str.Of(structName).Snake().String() + ".go"
				filterPath = filepath.Join("app/filters", folderPath, fileName)
			} else {
				structName := str.Of(tt.filterName).Studly().String()
				fileName := str.Of(structName).Snake().String() + ".go"
				filterPath = filepath.Join("app/filters", fileName)
			}
			assert.True(t, file.Exists(filterPath))

			// Verify registration in validation service provider
			if tt.expectSuccess {
				providerContent, err := file.GetContent(providerPath)
				assert.NoError(t, err)
				assert.Contains(t, providerContent, "app/filters")
			}
		})
	}
}

func TestFilterMakeCommand_RegistrationError(t *testing.T) {
	t.Run("handles error when bootstrap filters.go exists but WithFilters not registered", func(t *testing.T) {
		// Setup bootstrap directory with app.go but no WithFilters
		bootstrapDir := support.Config.Paths.Bootstrap
		appFile := filepath.Join(bootstrapDir, "app.go")
		filtersFile := filepath.Join(bootstrapDir, "filters.go")

		// Create bootstrap/app.go without WithFilters
		assert.NoError(t, file.PutContent(appFile, bootstrapAppFilter))
		// Create filters.go that shouldn't exist without WithFilters
		assert.NoError(t, file.PutContent(filtersFile, bootstrapFilters))
		defer func() {
			assert.NoError(t, file.Remove(bootstrapDir))
			assert.NoError(t, file.Remove("app"))
		}()

		// Setup mock context
		filterMakeCommand := &FilterMakeCommand{}
		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Argument(0).Return("FailFilter").Once()
		mockContext.EXPECT().OptionBool("force").Return(false).Once()
		mockContext.EXPECT().Success("Filter created successfully").Once()
		mockContext.EXPECT().Error(mock.MatchedBy(func(msg string) bool {
			return strings.Contains(msg, "filter register failed:")
		})).Once()

		// Execute command
		assert.NoError(t, filterMakeCommand.Handle(mockContext))
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
		filterMakeCommand := &FilterMakeCommand{}
		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Argument(0).Return("ErrorFilter").Once()
		mockContext.EXPECT().OptionBool("force").Return(false).Once()
		mockContext.EXPECT().Success("Filter created successfully").Once()
		mockContext.EXPECT().Error(mock.MatchedBy(func(msg string) bool {
			return strings.Contains(msg, "filter register failed:")
		})).Once()

		// Execute command
		assert.NoError(t, filterMakeCommand.Handle(mockContext))
	})
}
