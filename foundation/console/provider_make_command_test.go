package console

import (
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/console/command"
	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/modular"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/file"
)

type ProviderMakeCommandTestSuite struct {
	suite.Suite
}

func TestProviderMakeCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ProviderMakeCommandTestSuite))
}

func (s *ProviderMakeCommandTestSuite) TestSignature() {
	expected := "make:provider"
	cmd := &ProviderMakeCommand{}

	s.Require().Equal(expected, cmd.Signature())
}

func (s *ProviderMakeCommandTestSuite) TestDescription() {
	expected := "Create a new service provider class"
	cmd := &ProviderMakeCommand{}

	s.Require().Equal(expected, cmd.Description())
}

func (s *ProviderMakeCommandTestSuite) TestExtend() {
	cmd := &ProviderMakeCommand{}
	extend := cmd.Extend()

	s.Run("should return correct category", func() {
		expected := "make"
		s.Require().Equal(expected, extend.Category)
	})

	s.Run("should have correct number of flags", func() {
		s.Require().Len(extend.Flags, 2)
	})

	if len(extend.Flags) > 0 {
		s.Run("should have correctly configured BoolFlag", func() {
			flag, ok := extend.Flags[0].(*command.BoolFlag)
			if !ok {
				s.Fail("First flag is not BoolFlag (got type: %T)", extend.Flags[0])
			}

			testCases := []struct {
				name     string
				got      any
				expected any
			}{
				{"Name", flag.Name, "force"},
				{"Aliases", flag.Aliases, []string{"f"}},
				{"Usage", flag.Usage, "Create the provider even if it already exists"},
			}

			for _, tc := range testCases {
				if !reflect.DeepEqual(tc.got, tc.expected) {
					s.Require().Equal(tc.expected, tc.got)
				}
			}
		})
	}
}

func (s *ProviderMakeCommandTestSuite) TestHandle() {
	cmd := &ProviderMakeCommand{}
	mockContext := mocksconsole.NewContext(s.T())

	// Test empty name
	mockContext.EXPECT().Arguments().Return([]string{""}).Once()
	mockContext.EXPECT().Ask("Enter the provider name", mock.Anything).Return("", errors.New("the provider name cannot be empty")).Once()
	mockContext.EXPECT().Error("the provider name cannot be empty").Once()
	s.Nil(cmd.Handle(mockContext))

	// Test successful creation
	mockContext.EXPECT().Arguments().Return([]string{"UserServiceProvider"}).Once()
	mockContext.EXPECT().Option("module").Return("").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Provider created successfully").Once()
	s.NoError(cmd.Handle(mockContext))
	s.True(file.Exists("app/providers/user_service_provider.go"))

	// Test file already exists without force
	mockContext.EXPECT().Arguments().Return([]string{"UserServiceProvider"}).Once()
	mockContext.EXPECT().Option("module").Return("").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Error("the provider already exists. Use the --force or -f flag to overwrite").Once()
	s.Nil(cmd.Handle(mockContext))

	// Test file already exists with force
	mockContext.EXPECT().Arguments().Return([]string{"UserServiceProvider"}).Once()
	mockContext.EXPECT().Option("module").Return("").Once()
	mockContext.EXPECT().OptionBool("force").Return(true).Once()
	mockContext.EXPECT().Success("Provider created successfully").Once()
	s.NoError(cmd.Handle(mockContext))
	s.True(file.Exists("app/providers/user_service_provider.go"))

	// Test nested provider creation
	mockContext.EXPECT().Arguments().Return([]string{"auth/AuthServiceProvider"}).Once()
	mockContext.EXPECT().Option("module").Return("").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Provider created successfully").Once()
	s.NoError(cmd.Handle(mockContext))
	s.True(file.Exists("app/providers/auth/auth_service_provider.go"))
	s.True(file.Contain("app/providers/auth/auth_service_provider.go", "package auth"))
	s.True(file.Contain("app/providers/auth/auth_service_provider.go", "type AuthServiceProvider struct{}"))
	s.True(file.Contain("app/providers/auth/auth_service_provider.go", "func (r *AuthServiceProvider) Register(app foundation.Application) {"))

	// Clean up test files
	s.NoError(file.Remove("app"))
}

func (s *ProviderMakeCommandTestSuite) TestHandleMultipleNames() {
	s.T().Cleanup(func() {
		_ = file.Remove("app")
	})

	cmd := &ProviderMakeCommand{}
	mockContext := mocksconsole.NewContext(s.T())

	mockContext.EXPECT().Arguments().Return([]string{"AuthServiceProvider", "EventServiceProvider"}).Once()
	mockContext.EXPECT().Option("module").Return("").Times(2)
	mockContext.EXPECT().OptionBool("force").Return(false).Times(2)
	mockContext.EXPECT().Success("Provider created successfully").Times(2)
	s.NoError(cmd.Handle(mockContext))

	s.FileExists("app/providers/auth_service_provider.go")
	s.FileExists("app/providers/event_service_provider.go")
}

// blogModule sets its own directory, so it is a module of the application.
type blogModule struct{}

func (r *blogModule) Name() string { return "blog" }
func (r *blogModule) Path() string { return "modules/blog" }

// useBlogModule registers blogModule with a module.go on disk for the test.
func useBlogModule(t *testing.T) {
	registry := modular.NewRegistry("")
	assert.NoError(t, registry.Register(&blogModule{}))
	modular.SetDefault(registry)
	assert.NoError(t, file.PutContent("modules/blog/module.go", `package blog

type Module struct{}

func (m *Module) Name() string {
	return "blog"
}
`))

	t.Cleanup(func() {
		modular.SetDefault(nil)
		assert.NoError(t, file.Remove("modules"))
	})
}

func TestProviderMakeCommand_Module(t *testing.T) {
	useBlogModule(t)

	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Arguments().Return([]string{"BlogServiceProvider"}).Once()
	mockContext.EXPECT().Option("module").Return("blog").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Provider created successfully").Once()
	mockContext.EXPECT().Success("Provider registered successfully").Once()

	assert.NoError(t, NewProviderMakeCommand().Handle(mockContext))
	assert.True(t, file.Exists("modules/blog/providers/blog_service_provider.go"))

	content, err := file.GetContent("modules/blog/module.go")
	assert.NoError(t, err)
	assert.Contains(t, content, `/modules/blog/providers"`)
	assert.Contains(t, content, `func (m *Module) ServiceProviders() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&providers.BlogServiceProvider{},
	}
}`)
}

func TestProviderMakeCommand_ModuleRegistrationError(t *testing.T) {
	useBlogModule(t)
	// ServiceProviders() does not return a slice literal, so it cannot be edited.
	assert.NoError(t, file.PutContent("modules/blog/providers.go", `package blog

import "github.com/goravel/framework/contracts/foundation"

func (m *Module) ServiceProviders() []foundation.ServiceProvider {
	return nil
}
`))

	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Arguments().Return([]string{"BlogServiceProvider"}).Once()
	mockContext.EXPECT().Option("module").Return("blog").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Provider created successfully").Once()
	mockContext.EXPECT().Error(mock.MatchedBy(func(msg string) bool {
		return strings.Contains(msg, "BlogServiceProvider")
	})).Once()

	assert.NoError(t, NewProviderMakeCommand().Handle(mockContext))
	assert.True(t, file.Exists("modules/blog/providers/blog_service_provider.go"))
}

func TestProviderMakeCommand_BootstrapSetup(t *testing.T) {
	bootstrapDir := support.Config.Paths.Bootstrap
	assert.NoError(t, file.PutContent(filepath.Join(bootstrapDir, "app.go"), `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().Start()
}
`))
	t.Cleanup(func() {
		assert.NoError(t, file.Remove(bootstrapDir))
		assert.NoError(t, file.Remove("app"))
	})

	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Arguments().Return([]string{"UserServiceProvider"}).Once()
	mockContext.EXPECT().Option("module").Return("").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Provider created successfully").Once()
	mockContext.EXPECT().Success("Provider registered successfully").Once()

	assert.NoError(t, NewProviderMakeCommand().Handle(mockContext))
	assert.True(t, file.Contain(filepath.Join(bootstrapDir, "providers.go"), "&providers.UserServiceProvider{}"))
}
