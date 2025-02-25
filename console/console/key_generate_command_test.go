package console

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

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
	// Linux and Windows error message are different
	mockContext.EXPECT().Error(mock.MatchedBy(func(s string) bool {
		return strings.Contains(s, "open .env:")
	})).Once()

	assert.False(t, file.Exists(support.EnvPath))

	assert.Nil(t, keyGenerateCommand.Handle(mockContext))

	err := file.PutContent(support.EnvPath, "APP_KEY=12345\n")
	assert.Nil(t, err)

	mockContext.EXPECT().Success("Application key set successfully").Once()
	assert.Nil(t, keyGenerateCommand.Handle(mockContext))
	assert.True(t, file.Exists(support.EnvPath))
	env, err := os.ReadFile(support.EnvPath)
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

	env, err = os.ReadFile(support.EnvPath)
	assert.Nil(t, err)
	assert.True(t, len(env) > 10)
	assert.Nil(t, file.Remove(support.EnvPath))
}

func TestKeyGenerateCommandWithCustomEnvFile(t *testing.T) {
	support.EnvPath = "config.conf"

	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("app.env").Return("local").Twice()
	mockConfig.EXPECT().GetString("app.key").Return("12345").Once()

	keyGenerateCommand := NewKeyGenerateCommand(mockConfig)
	mockContext := mocksconsole.NewContext(t)
	// Linux and Windows error message are different
	mockContext.EXPECT().Error(mock.MatchedBy(func(s string) bool {
		return strings.Contains(s, "open config.conf:")
	})).Once()
	assert.False(t, file.Exists("config.conf"))

	assert.Nil(t, keyGenerateCommand.Handle(mockContext))

	err := file.PutContent("config.conf", "APP_KEY=12345\n")
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
