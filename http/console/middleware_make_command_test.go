package console

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/file"
)

func TestMiddlewareMakeCommand(t *testing.T) {
	middlewareMakeCommand := &MiddlewareMakeCommand{}
	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Argument(0).Return("").Once()
	mockContext.EXPECT().Ask("Enter the middleware name", mock.Anything).Return("", errors.New("the middleware name cannot be empty")).Once()
	mockContext.EXPECT().Error("the middleware name cannot be empty").Once()
	assert.NoError(t, middlewareMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("VerifyCsrfToken").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Middleware created successfully").Once()
	assert.NoError(t, middlewareMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/http/middleware/verify_csrf_token.go"))

	mockContext.EXPECT().Argument(0).Return("VerifyCsrfToken").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Error("the middleware already exists. Use the --force or -f flag to overwrite").Once()
	assert.NoError(t, middlewareMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("User/Auth").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Middleware created successfully").Once()
	assert.NoError(t, middlewareMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/http/middleware/User/auth.go"))
	assert.True(t, file.Contain("app/http/middleware/User/auth.go", "package User"))
	assert.True(t, file.Contain("app/http/middleware/User/auth.go", "func Auth() http.Middleware {"))
	assert.Nil(t, file.Remove("app"))
}

func TestMiddlewareMakeCommand_WithBootstrapSetup(t *testing.T) {
	middlewareMakeCommand := &MiddlewareMakeCommand{}
	bootstrapPath := filepath.Join("bootstrap", "app.go")

	// Ensure clean state before test
	defer func() {
		assert.Nil(t, file.Remove("bootstrap"))
		assert.Nil(t, file.Remove("app"))
	}()

	// Create bootstrap/app.go with foundation.Setup()
	bootstrapContent := `package bootstrap

import (
	"github.com/goravel/framework/foundation"
)

func Boot() {
	foundation.Setup().Run()
}
`
	assert.NoError(t, file.PutContent(bootstrapPath, bootstrapContent))

	// Create mock context
	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Argument(0).Return("Auth").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Middleware created successfully").Once()
	mockContext.EXPECT().Success("Middleware registered successfully").Once()

	// Execute
	err := middlewareMakeCommand.Handle(mockContext)

	// Assert
	assert.Nil(t, err)

	// Verify the middleware file was created
	middlewarePath := filepath.Join("app", "http", "middleware", "auth.go")
	assert.True(t, file.Exists(middlewarePath))
	assert.True(t, file.Contain(middlewarePath, "package middleware"))
	assert.True(t, file.Contain(middlewarePath, "func Auth() http.Middleware {"))

	// Verify bootstrap/app.go was modified with AddMiddleware
	bootstrapContent, readErr := file.GetContent(bootstrapPath)
	assert.NoError(t, readErr)
	expectedContent := `package bootstrap

import (
	"github.com/goravel/framework/app/http/middleware"
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
)

func Boot() {
	foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
			handler.Append(
				middleware.Auth(),
			)
		}).Run()
}
`
	assert.Equal(t, expectedContent, bootstrapContent)
}

func TestMiddlewareMakeCommand_WithBootstrapSetupRegistrationFailed(t *testing.T) {
	middlewareMakeCommand := &MiddlewareMakeCommand{}
	bootstrapPath := filepath.Join("bootstrap", "app.go")

	// Ensure clean state before test
	defer func() {
		assert.Nil(t, file.Remove("bootstrap"))
		assert.Nil(t, file.Remove("app"))
	}()

	// Create bootstrap/app.go with invalid syntax that will cause AddMiddleware to fail parsing
	// It contains foundation.Setup() so IsBootstrapSetup() returns true,
	// but has invalid Go syntax that will cause the parser to fail
	bootstrapContent := `package bootstrap

import (
	"github.com/goravel/framework/foundation"
)

func Boot() {
	foundation.Setup().
}
`
	assert.NoError(t, file.PutContent(bootstrapPath, bootstrapContent))

	// Create mock context
	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Argument(0).Return("Auth").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Middleware created successfully").Once()
	mockContext.EXPECT().Error(mock.MatchedBy(func(msg string) bool {
		return assert.Contains(t, msg, "failed to register middleware 'Auth'")
	})).Once()

	// Execute
	err := middlewareMakeCommand.Handle(mockContext)

	// Assert - should return nil (errors are handled via ctx.Error)
	assert.Nil(t, err)

	// Verify the middleware file was created (creation happens before registration)
	middlewarePath := filepath.Join("app", "http", "middleware", "auth.go")
	assert.True(t, file.Exists(middlewarePath))
}
