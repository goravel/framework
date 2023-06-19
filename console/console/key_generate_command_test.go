package console

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	configmock "github.com/goravel/framework/contracts/config/mocks"
	consolemocks "github.com/goravel/framework/contracts/console/mocks"
	"github.com/goravel/framework/support/file"
)

func TestKeyGenerateCommand(t *testing.T) {
	mockConfig := &configmock.Config{}
	mockConfig.On("GetString", "app.env").Return("local").Twice()
	mockConfig.On("GetString", "app.key").Return("12345").Twice()

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
	input := "no\n"

	reader, writer, err := os.Pipe()
	assert.Nil(t, err)
	originalStdin := os.Stdin
	defer func() { os.Stdin = originalStdin }()
	os.Stdin = reader
	go func() {
		defer func(writer *os.File) {
			assert.Nil(t, writer.Close())
		}(writer)
		_, err = io.WriteString(writer, input)
		assert.Nil(t, err)
	}()

	assert.Nil(t, keyGenerateCommand.Handle(mockContext))
	env, err = os.ReadFile(".env")
	assert.Nil(t, err)
	assert.True(t, len(env) > 10)
	assert.True(t, file.Remove(".env"))

	mockConfig.AssertExpectations(t)
}
