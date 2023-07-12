package console

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	configmock "github.com/goravel/framework/contracts/config/mocks"
	consolemocks "github.com/goravel/framework/contracts/console/mocks"
	"github.com/goravel/framework/support"
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

	reader, writer, err := os.Pipe()
	assert.Nil(t, err)
	originalStdin := os.Stdin
	defer func() { os.Stdin = originalStdin }()
	os.Stdin = reader
	go func() {
		defer writer.Close()
		_, err = writer.Write([]byte("no\n"))
		assert.Nil(t, err)
	}()

	assert.Nil(t, keyGenerateCommand.Handle(mockContext))
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

	reader, writer, err := os.Pipe()
	assert.Nil(t, err)
	originalStdin := os.Stdin
	defer func() { os.Stdin = originalStdin }()
	os.Stdin = reader
	go func() {
		defer writer.Close()
		_, err = writer.Write([]byte("no\n"))
		assert.Nil(t, err)
	}()

	assert.Nil(t, keyGenerateCommand.Handle(mockContext))
	env, err = os.ReadFile("config.conf")
	assert.Nil(t, err)
	assert.True(t, len(env) > 10)
	assert.Nil(t, file.Remove("config.conf"))

	support.EnvPath = ".env"

	mockConfig.AssertExpectations(t)
}
