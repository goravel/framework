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

func TestJwtSecretCommand(t *testing.T) {
	mockConfig := &configmock.Config{}
	mockConfig.On("GetString", "jwt.secret").Return("").Twice()

	jwtSecretCommand := NewJwtSecretCommand(mockConfig)
	mockContext := &consolemocks.Context{}

	assert.False(t, file.Exists(".env"))
	err := file.Create(".env", "JWT_SECRET=\n")
	assert.Nil(t, err)

	assert.Nil(t, jwtSecretCommand.Handle(mockContext))

	assert.True(t, file.Exists(".env"))
	env, err := os.ReadFile(".env")
	assert.Nil(t, err)
	assert.True(t, len(env) > 10)
	assert.Nil(t, file.Remove(".env"))

	mockConfig.AssertExpectations(t)
}

func TestJwtSecretCommandWithCustomEnvFile(t *testing.T) {
	support.EnvPath = "config.conf"

	mockConfig := &configmock.Config{}
	mockConfig.On("GetString", "jwt.secret").Return("").Twice()

	jwtSecretCommand := NewJwtSecretCommand(mockConfig)
	mockContext := &consolemocks.Context{}

	assert.False(t, file.Exists("config.conf"))
	err := file.Create("config.conf", "JWT_SECRET=\n")
	assert.Nil(t, err)

	assert.Nil(t, jwtSecretCommand.Handle(mockContext))

	assert.True(t, file.Exists("config.conf"))
	env, err := os.ReadFile("config.conf")
	assert.Nil(t, err)
	assert.True(t, len(env) > 10)
	assert.Nil(t, file.Remove("config.conf"))

	support.EnvPath = ".env"

	mockConfig.AssertExpectations(t)
}
