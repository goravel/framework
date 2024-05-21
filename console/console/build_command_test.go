package console

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/color"
)

func TestBuildCommand(t *testing.T) {
	mockConfig := &mocksconfig.Config{}
	mockConfig.On("GetString", "app.env").Return("local").Once()

	newBuildCommand := NewBuildCommand(mockConfig)
	mockContext := &mocksconsole.Context{}
	mockContext.On("Option", "system").Return("invalidSystem").Once()

	assert.NotNil(t, newBuildCommand.Handle(mockContext))

	mockConfig.On("GetString", "app.env").Return("production").Once()
	mockContext.On("Confirm", "Do you really wish to run this command?").Return(false, nil).Once()

	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, newBuildCommand.Handle(mockContext))
	}), "Command cancelled!")

	mockConfig.On("GetString", "app.env").Return("production").Once()
	mockContext.On("Confirm", "Do you really wish to run this command?").Return(false, errors.New("error")).Once()

	assert.NotContains(t, color.CaptureOutput(func(w io.Writer) {
		assert.EqualError(t, newBuildCommand.Handle(mockContext), "error")
	}), "Command cancelled!")
}
