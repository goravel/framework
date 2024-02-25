package console

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
)

func TestBuildCommand(t *testing.T) {
	mockConfig := &mocksconfig.Config{}
	mockConfig.On("GetString", "app.env").Return("local").Once()

	newBuildCommand := NewBuildCommand(mockConfig)
	mockContext := &mocksconsole.Context{}
	mockContext.On("Option", "system").Return("invalidSystem").Once()

	assert.NotNil(t, newBuildCommand.Handle(mockContext))

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
