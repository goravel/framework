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

func TestKeyGenerateCommand(t *testing.T) {
	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("app.env").Return("local").Twice()
	mockConfig.EXPECT().GetString("app.key").Return("12345").Once()

	keyGenerateCommand := NewKeyGenerateCommand(mockConfig)
	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Error("open .env: no such file or directory").Once()

	assert.False(t, file.Exists(".env"))

	assert.Nil(t, keyGenerateCommand.Handle(mockContext))

	err := file.Create(".env", "APP_KEY=12345\n")
	assert.Nil(t, err)

	mockContext.EXPECT().Success("Application key set successfully").Once()
	assert.Nil(t, keyGenerateCommand.Handle(mockContext))
	assert.True(t, file.Exists(".env"))
	env, err := os.ReadFile(".env")
	assert.Nil(t, err)
	assert.True(t, len(env) > 10)

	mockConfig.EXPECT().GetString("app.env").Return("production").Once()
	mockContext.EXPECT().Confirm("Do you really wish to run this command?").Return(false, nil).Once()
	mockContext.EXPECT().Warning("Command cancelled!").Once()
	assert.Nil(t, keyGenerateCommand.Handle(mockContext))

	mockConfig.EXPECT().GetString("app.env").Return("production").Once()
	mockContext.EXPECT().Confirm("Do you really wish to run this command?").Return(false, assert.AnError).Once()
	mockContext.EXPECT().Error(assert.AnError.Error()).Once()
	assert.Nil(t, keyGenerateCommand.Handle(mockContext))

	env, err = os.ReadFile(".env")
	assert.Nil(t, err)
	assert.True(t, len(env) > 10)
	assert.Nil(t, file.Remove(".env"))
}

func TestKeyGenerateCommandWithCustomEnvFile(t *testing.T) {
	support.EnvPath = "config.conf"

	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("app.env").Return("local").Twice()
	mockConfig.EXPECT().GetString("app.key").Return("12345").Once()

	keyGenerateCommand := NewKeyGenerateCommand(mockConfig)
	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Error("open config.conf: no such file or directory").Once()
	assert.False(t, file.Exists("config.conf"))

	assert.Nil(t, keyGenerateCommand.Handle(mockContext))

	err := file.Create("config.conf", "APP_KEY=12345\n")
	assert.Nil(t, err)

	mockContext.EXPECT().Success("Application key set successfully").Once()
	assert.Nil(t, keyGenerateCommand.Handle(mockContext))
	assert.True(t, file.Exists("config.conf"))
	env, err := os.ReadFile("config.conf")
	assert.Nil(t, err)
	assert.True(t, len(env) > 10)

	mockConfig.EXPECT().GetString("app.env").Return("production").Once()
	mockContext.EXPECT().Confirm("Do you really wish to run this command?").Return(false, nil).Once()
	mockContext.EXPECT().Warning("Command cancelled!").Once()
	assert.Nil(t, keyGenerateCommand.Handle(mockContext))

	env, err = os.ReadFile("config.conf")
	assert.Nil(t, err)
	assert.True(t, len(env) > 10)
	assert.Nil(t, file.Remove("config.conf"))

	support.EnvPath = ".env"
}
