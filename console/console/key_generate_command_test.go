package console

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	configmock "github.com/goravel/framework/mocks/config"
	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/file"
)

func TestKeyGenerateCommand(t *testing.T) {
	mockConfig := &configmock.Config{}
	mockConfig.On("GetString", "app.env").Return("local").Twice()
	mockConfig.On("GetString", "app.key").Return("12345").Once()

	keyGenerateCommand := NewKeyGenerateCommand(mockConfig)
	mockContext := &consolemocks.Context{}

	assert.False(t, file.Exists(".env"))

	assert.Nil(t, keyGenerateCommand.Handle(mockContext))

	err := file.Create(".env", "APP_KEY=12345\n")
	assert.Nil(t, err)

	assert.Nil(t, keyGenerateCommand.Handle(mockContext))
	assert.True(t, file.Exists(".env"))
	env, err := os.ReadFile(".env")
	assert.Nil(t, err)
	assert.True(t, len(env) > 10)

	mockConfig.On("GetString", "app.env").Return("production").Once()
	mockContext.On("Confirm", "Do you really wish to run this command?").Return(false, nil).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, keyGenerateCommand.Handle(mockContext))
	}), "Command cancelled!")

	mockConfig.On("GetString", "app.env").Return("production").Once()
	mockContext.On("Confirm", "Do you really wish to run this command?").Return(false, errors.New("error")).Once()
	assert.NotContains(t, color.CaptureOutput(func(w io.Writer) {
		assert.EqualError(t, keyGenerateCommand.Handle(mockContext), "error")
	}), "Command cancelled!")

	env, err = os.ReadFile(".env")
	assert.Nil(t, err)
	assert.True(t, len(env) > 10)
	assert.Nil(t, file.Remove(".env"))

	mockConfig.AssertExpectations(t)
}

func TestKeyGenerateCommandWithCustomEnvFile(t *testing.T) {
	support.EnvPath = "config.conf"

	mockConfig := &configmock.Config{}
	mockConfig.On("GetString", "app.env").Return("local").Twice()
	mockConfig.On("GetString", "app.key").Return("12345").Once()

	keyGenerateCommand := NewKeyGenerateCommand(mockConfig)
	mockContext := &consolemocks.Context{}

	assert.False(t, file.Exists("config.conf"))

	assert.Nil(t, keyGenerateCommand.Handle(mockContext))

	err := file.Create("config.conf", "APP_KEY=12345\n")
	assert.Nil(t, err)

	assert.Nil(t, keyGenerateCommand.Handle(mockContext))
	assert.True(t, file.Exists("config.conf"))
	env, err := os.ReadFile("config.conf")
	assert.Nil(t, err)
	assert.True(t, len(env) > 10)

	mockConfig.On("GetString", "app.env").Return("production").Once()
	mockContext.On("Confirm", "Do you really wish to run this command?").Return(false, nil).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, keyGenerateCommand.Handle(mockContext))
	}), "Command cancelled!")

	env, err = os.ReadFile("config.conf")
	assert.Nil(t, err)
	assert.True(t, len(env) > 10)
	assert.Nil(t, file.Remove("config.conf"))

	support.EnvPath = ".env"

	mockConfig.AssertExpectations(t)
}
