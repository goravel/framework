package console

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	configmock "github.com/goravel/framework/mocks/config"
	consolemocks "github.com/goravel/framework/mocks/console"
)

func TestBuildCommand(t *testing.T) {
	mockConfig := &configmock.Config{}
	mockConfig.On("GetString", "app.env").Return("local").Once()

	newBuildCommand := NewBuildCommand(mockConfig)

	assert.Equal(t, newBuildCommand.Signature(), "build")
	assert.Equal(t, newBuildCommand.Description(), "Build the application")

	mockContext := &consolemocks.Context{}
	mockContext.On("Option", "system").Return("linux").Once()

	assert.Nil(t, newBuildCommand.Handle(mockContext))

	mockConfig.On("GetString", "app.env").Return("production").Once()
	mockContext.On("Option", "system").Return("linux").Once()

	reader, writer, err := os.Pipe()
	assert.Nil(t, err)
	originalStdin := os.Stdin
	defer func() { os.Stdin = originalStdin }()
	os.Stdin = reader
	go func() {
		defer writer.Close()
		_, err = writer.Write([]byte("yes\n"))
		assert.Nil(t, err)
	}()

	assert.Nil(t, newBuildCommand.Handle(mockContext))

	mockConfig.On("GetString", "app.env").Return("production").Once()
	mockContext.On("Option", "system").Return("linux").Once()

	reader, writer, err = os.Pipe()
	assert.Nil(t, err)
	originalStdin = os.Stdin
	defer func() { os.Stdin = originalStdin }()
	os.Stdin = reader
	go func() {
		defer writer.Close()
		_, err = writer.Write([]byte("no\n"))
		assert.Nil(t, err)
	}()

	assert.Nil(t, newBuildCommand.Handle(mockContext))
}
