package console

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/file"
)

func TestJwtSecretCommand(t *testing.T) {
	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("jwt.secret").Return("").Twice()

	jwtSecretCommand := NewJwtSecretCommand(mockConfig)
	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Success("Jwt Secret set successfully").Once()

	assert.False(t, file.Exists(support.EnvFilePath))
	err := file.PutContent(support.EnvFilePath, "JWT_SECRET=\n")
	assert.NoError(t, err)

	assert.NoError(t, jwtSecretCommand.Handle(mockContext))

	assert.True(t, file.Exists(support.EnvFilePath))
	env, err := os.ReadFile(support.EnvFilePath)
	assert.NoError(t, err)
	assert.True(t, len(env) > 10)
	assert.NoError(t, file.Remove(support.EnvFilePath))
}

func TestJwtSecretCommandWithCustomEnvFile(t *testing.T) {
	support.EnvFilePath = "config.conf"

	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("jwt.secret").Return("").Twice()

	jwtSecretCommand := NewJwtSecretCommand(mockConfig)
	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Success("Jwt Secret set successfully").Once()

	assert.False(t, file.Exists("config.conf"))
	err := file.PutContent("config.conf", "JWT_SECRET=\n")
	assert.NoError(t, err)

	assert.NoError(t, jwtSecretCommand.Handle(mockContext))

	assert.True(t, file.Exists("config.conf"))
	env, err := os.ReadFile("config.conf")
	assert.NoError(t, err)
	assert.True(t, len(env) > 10)
	assert.NoError(t, file.Remove("config.conf"))

	support.EnvFilePath = ".env"
}
