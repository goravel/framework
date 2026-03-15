package console

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	mocksconsole "github.com/goravel/framework/mocks/console"
)

func TestAgentsInstallCommand(t *testing.T) {
	var (
		mockContext    *mocksconsole.Context
		installCommand *AgentsInstallCommand
	)

	beforeEach := func() {
		mockContext = mocksconsole.NewContext(t)
		installCommand = &AgentsInstallCommand{}
	}

	cleanup := func() {
		os.RemoveAll(".ai")
		os.Remove("AGENTS.md")
	}

	filePaths := []string{"AGENTS.md", "prompt/route.md"}

	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "Happy path - install all files",
			setup: func() {
				installCommand.treeFetcher = func(branch string) ([]string, error) {
					return filePaths, nil
				}
				installCommand.fetcher = func(branch, path string) ([]byte, error) {
					return []byte("# " + path), nil
				}

				mockContext.EXPECT().Option("version").Return("v1.17").Once()
				mockContext.EXPECT().OptionBool("force").Return(true).Once()
				mockContext.EXPECT().Option("file").Return("").Once()
				mockContext.EXPECT().Info("Installed 2 file(s) for version v1.17.").Once()
			},
		},
		{
			name: "Happy path - falls back to master when version branch has no agent files",
			setup: func() {
				installCommand.treeFetcher = func(branch string) ([]string, error) {
					if branch == docsFallbackBranch {
						return filePaths, nil
					}
					return nil, nil
				}
				installCommand.fetcher = func(branch, path string) ([]byte, error) {
					return []byte("# " + path), nil
				}

				mockContext.EXPECT().Option("version").Return("v1.99").Once()
				mockContext.EXPECT().OptionBool("force").Return(true).Once()
				mockContext.EXPECT().Option("file").Return("").Once()
				mockContext.EXPECT().Info("Installed 2 file(s) for version v1.99.").Once()
			},
		},
		{
			name: "Sad path - no files on branch or master",
			setup: func() {
				installCommand.treeFetcher = func(branch string) ([]string, error) {
					return nil, nil
				}

				mockContext.EXPECT().Option("version").Return("v9.99").Once()
				mockContext.EXPECT().Error("No agent files found for version v9.99. Check https://github.com/goravel/docs").Once()
			},
		},
		{
			name: "Sad path - unsupported version",
			setup: func() {
				mockContext.EXPECT().Option("version").Return("v1.16").Once()
				mockContext.EXPECT().Error("agent files are only available for Goravel v1.17 and above (got v1.16)").Once()
			},
		},
		{
			name: "Sad path - tree fetch error",
			setup: func() {
				installCommand.treeFetcher = func(branch string) ([]string, error) {
					return nil, errors.New("network error")
				}

				mockContext.EXPECT().Option("version").Return("v1.17").Once()
				mockContext.EXPECT().Error("network error").Once()
			},
		},
		{
			name: "Sad path - file filter no match",
			setup: func() {
				installCommand.treeFetcher = func(branch string) ([]string, error) {
					return filePaths, nil
				}

				mockContext.EXPECT().Option("version").Return("v1.17").Once()
				mockContext.EXPECT().OptionBool("force").Return(true).Once()
				mockContext.EXPECT().Option("file").Return("nonexistent").Once()
				mockContext.EXPECT().Error("No file matching 'nonexistent' found in remote repository.").Once()
			},
		},
		{
			name: "Sad path - existing install, user cancels",
			setup: func() {
				installCommand.treeFetcher = func(branch string) ([]string, error) {
					return filePaths, nil
				}

				os.MkdirAll(".ai", 0755)
				os.WriteFile(versionFilePath, []byte(`{"version":"v1.16","files":{}}`), 0644)

				mockContext.EXPECT().Option("version").Return("v1.17").Once()
				mockContext.EXPECT().OptionBool("force").Return(false).Once()
				mockContext.EXPECT().Confirm("Agent files are already installed. Overwrite?").Return(false).Once()
				mockContext.EXPECT().Warning("Cancelled.").Once()
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			cleanup()
			test.setup()
			defer cleanup()

			assert.Nil(t, installCommand.Handle(mockContext))
		})
	}
}
