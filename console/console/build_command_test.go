package console

import (
	"errors"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/console"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksprocess "github.com/goravel/framework/mocks/process"
)

func TestBuildCommand(t *testing.T) {
	var (
		mockConfig   *mocksconfig.Config
		mockContext  *mocksconsole.Context
		mockProcess  *mocksprocess.Process
		mockResult   *mocksprocess.Result
		buildCommand *BuildCommand
	)

	beforeEach := func() {
		mockConfig = mocksconfig.NewConfig(t)
		mockContext = mocksconsole.NewContext(t)
		mockProcess = mocksprocess.NewProcess(t)
		mockResult = mocksprocess.NewResult(t)
		buildCommand = NewBuildCommand(mockConfig, mockProcess)
	}

	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "Happy path",
			setup: func() {
				mockConfig.EXPECT().GetString("app.env").Return("local").Once()
				mockContext.EXPECT().Option("os").Return("linux").Once()
				mockContext.EXPECT().Option("arch").Return("amd64").Once()
				mockContext.EXPECT().Option("name").Return("").Once()
				mockContext.EXPECT().OptionBool("static").Return(true).Once()
				mockProcess.EXPECT().Env(map[string]string{
					"CGO_ENABLED": "0",
					"GOOS":        "linux",
					"GOARCH":      "amd64",
				}).Return(mockProcess).Once()
				mockProcess.EXPECT().WithSpinner("Building...").Return(mockProcess).Once()
				mockProcess.EXPECT().Run("go build -ldflags -extldflags -static .").Return(mockResult).Once()
				mockResult.EXPECT().Failed().Return(false).Once()
				mockContext.EXPECT().Info("Built successfully.").Once()
			},
		},
		{
			name: "Sad path - env is prod and confirm false",
			setup: func() {
				mockConfig.EXPECT().GetString("app.env").Return("production").Once()
				mockContext.EXPECT().Warning("**************************************").Once()
				mockContext.EXPECT().Warning("*     Application In Production!     *").Once()
				mockContext.EXPECT().Warning("**************************************").Once()
				mockContext.EXPECT().Confirm("Do you really wish to run this command?").Return(false).Once()
				mockContext.EXPECT().Warning("Command cancelled!").Once()
			},
		},
		{
			name: "Sad path - os is empty and choice error",
			setup: func() {
				mockConfig.EXPECT().GetString("app.env").Return("local").Once()
				mockContext.EXPECT().Option("os").Return("").Once()
				mockContext.EXPECT().Choice("Select target os", []console.Choice{
					{Key: "Linux", Value: "linux"},
					{Key: "Windows", Value: "windows"},
					{Key: "Darwin", Value: "darwin"},
				}, console.ChoiceOption{Default: runtime.GOOS}).Return("", errors.New("error")).Once()
				mockContext.EXPECT().Error("Select target os error: error").Once()
			},
		},
		{
			name: "Sad path - os/arch is invalid",
			setup: func() {
				mockConfig.EXPECT().GetString("app.env").Return("local").Once()
				mockContext.EXPECT().Option("os").Return("invalid").Once()
				mockContext.EXPECT().Option("arch").Return("invalid").Once()
				mockContext.EXPECT().Option("name").Return("").Once()
				mockContext.EXPECT().OptionBool("static").Return(true).Once()
				mockProcess.EXPECT().Env(map[string]string{
					"CGO_ENABLED": "0",
					"GOOS":        "invalid",
					"GOARCH":      "invalid",
				}).Return(mockProcess).Once()
				mockProcess.EXPECT().WithSpinner("Building...").Return(mockProcess).Once()
				mockProcess.EXPECT().Run("go build -ldflags -extldflags -static .").Return(mockResult).Once()
				mockResult.EXPECT().Failed().Return(true).Once()
				mockResult.EXPECT().Error().Return(errors.New("go: unsupported GOOS/GOARCH pair invalid/invalid")).Once()
				mockContext.EXPECT().Error("go: unsupported GOOS/GOARCH pair invalid/invalid").Once()
			},
		},
		{
			name: "Sad path - run returns error",
			setup: func() {
				mockConfig.EXPECT().GetString("app.env").Return("local").Once()
				mockContext.EXPECT().Option("os").Return("linux").Once()
				mockContext.EXPECT().Option("arch").Return("amd64").Once()
				mockContext.EXPECT().Option("name").Return("").Once()
				mockContext.EXPECT().OptionBool("static").Return(true).Once()
				mockProcess.EXPECT().Env(map[string]string{
					"CGO_ENABLED": "0",
					"GOOS":        "linux",
					"GOARCH":      "amd64",
				}).Return(mockProcess).Once()
				mockProcess.EXPECT().WithSpinner("Building...").Return(mockProcess).Once()
				mockProcess.EXPECT().Run("go build -ldflags -extldflags -static .").Return(mockResult).Once()
				mockResult.EXPECT().Failed().Return(true).Once()
				mockResult.EXPECT().Error().Return(errors.New("error")).Once()
				mockContext.EXPECT().Error("error").Once()
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			test.setup()

			assert.Nil(t, buildCommand.Handle(mockContext))

			mockConfig.AssertExpectations(t)
		})
	}
}

func TestGenerateCommand(t *testing.T) {
	tests := []struct {
		description string
		name        string
		static      bool
		expected    string
	}{
		{
			description: "Generate command with static and name",
			name:        "test",
			static:      true,
			expected:    "go build -ldflags -extldflags -static -o test .",
		},
		{
			description: "Generate command with static without name",
			static:      true,
			expected:    "go build -ldflags -extldflags -static .",
		},
		{
			description: "Generate command without static with name",
			name:        "test",
			static:      false,
			expected:    "go build -o test .",
		},
		{
			description: "Generate command without static and name",
			static:      false,
			expected:    "go build .",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result := generateCommand(test.name, test.static)

			assert.Equal(t, test.expected, result)
		})
	}
}
