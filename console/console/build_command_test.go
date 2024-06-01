package console

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/console"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
)

func TestBuildCommand(t *testing.T) {
	var (
		mockConfig   *mocksconfig.Config
		mockContext  *mocksconsole.Context
		buildCommand *BuildCommand
	)

	beforeEach := func() {
		mockConfig = &mocksconfig.Config{}
		mockContext = &mocksconsole.Context{}
		buildCommand = NewBuildCommand(mockConfig)
	}

	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "env is prod and confirm error",
			setup: func() {
				mockConfig.EXPECT().GetString("app.env").Return("production").Once()
				mockContext.EXPECT().Warning("**************************************").Once()
				mockContext.EXPECT().Warning("*     Application In Production!     *").Once()
				mockContext.EXPECT().Warning("**************************************").Once()
				mockContext.EXPECT().Confirm("Do you really wish to run this command?").Return(false, errors.New("error")).Once()
				mockContext.EXPECT().Error("Confirm error: error").Once()
			},
		},
		{
			name: "env is prod and confirm false",
			setup: func() {
				mockConfig.EXPECT().GetString("app.env").Return("production").Once()
				mockContext.EXPECT().Warning("**************************************").Once()
				mockContext.EXPECT().Warning("*     Application In Production!     *").Once()
				mockContext.EXPECT().Warning("**************************************").Once()
				mockContext.EXPECT().Confirm("Do you really wish to run this command?").Return(false, nil).Once()
				mockContext.EXPECT().Warning("Command cancelled!").Once()
			},
		},
		{
			name: "system is empty and choice error",
			setup: func() {
				mockConfig.EXPECT().GetString("app.env").Return("local").Once()
				mockContext.EXPECT().Option("system").Return("").Once()
				mockContext.EXPECT().Choice("Select target system os", []console.Choice{
					{Key: "Linux", Value: "linux"},
					{Key: "Darwin", Value: "windows"},
					{Key: "Windows", Value: "darwin"},
				}).Return("", errors.New("error")).Once()
				mockContext.EXPECT().Error("Select target system error: error").Once()
			},
		},
		{
			name: "system is invalid",
			setup: func() {
				mockConfig.EXPECT().GetString("app.env").Return("local").Once()
				mockContext.EXPECT().Option("system").Return("invalid").Once()
				mockContext.EXPECT().Error("Invalid system 'invalid' specified. Allowed values are: [linux windows darwin]").Once()
			},
		},
		{
			name: "system is valid",
			setup: func() {
				mockConfig.EXPECT().GetString("app.env").Return("local").Once()
				mockContext.EXPECT().Option("system").Return("linux").Once()
				mockContext.EXPECT().Spinner("Building...", mock.Anything).Return().Once()
				mockContext.EXPECT().Info("Built successfully.").Once()
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			test.setup()

			buildCommand.Handle(mockContext)

			mockConfig.AssertExpectations(t)
		})
	}

	//mockConfig := &mocksconfig.Config{}
	//mockConfig.On("GetString", "app.env").Return("local").Once()
	//
	//newBuildCommand := NewBuildCommand(mockConfig)
	//mockContext := &mocksconsole.Context{}
	//mockContext.On("Option", "system").Return("invalidSystem").Once()
	//mockContext.EXPECT().Error()
	//	assert.Nil(t, newBuildCommand.Handle(mockContext))
	//
	//mockConfig.On("GetString", "app.env").Return("production").Once()
	//mockContext.On("Confirm", "Do you really wish to run this command?").Return(false, nil).Once()
	//
	//assert.Equal(t, color.CaptureOutput(func(w io.Writer) {
	//	assert.Nil(t, newBuildCommand.Handle(mockContext))
	//}), "Command cancelled!")
	//
	//mockConfig.On("GetString", "app.env").Return("production").Once()
	//mockContext.On("Confirm", "Do you really wish to run this command?").Return(false, errors.New("error")).Once()
	//
	//assert.Equal(t, color.CaptureOutput(func(w io.Writer) {
	//	assert.Nil(t, newBuildCommand.Handle(mockContext), "error")
	//}), "Confirm error: error")
	//
	//mockConfig.AssertExpectations(t)
	//mockContext.AssertExpectations(t)
}

func TestGenerateCommand(t *testing.T) {
	tests := []struct {
		description string
		name        string
		static      bool
		expected    []string
	}{
		{
			description: "Generate command with static and name",
			name:        "test",
			static:      true,
			expected:    []string{"go", "build", "-ldflags", "-extldflags -static", "-o", "test", "."},
		},
		{
			description: "Generate command with static without name",
			name:        "",
			static:      true,
			expected:    []string{"go", "build", "-ldflags", "-extldflags -static", "."},
		},
		{
			description: "Generate command without static with name",
			name:        "test",
			static:      false,
			expected:    []string{"go", "build", "-o", "test", "."},
		},
		{
			description: "Generate command without static and name",
			name:        "",
			static:      false,
			expected:    []string{"go", "build", "."},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result := generateCommand(test.name, test.static)

			assert.Equal(t, test.expected, result)
		})
	}
}
