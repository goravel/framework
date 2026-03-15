package console

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	mocksconsole "github.com/goravel/framework/mocks/console"
)

func TestAgentsUpdateCommandMissingVersionFile(t *testing.T) {
	os.RemoveAll(".ai")

	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Error("No .ai/.version found. Run agents:install first.").Once()

	cmd := &AgentsUpdateCommand{
		treeFetcher: func(branch string) ([]string, error) { return nil, nil },
		fetcher:     func(branch, path string) ([]byte, error) { return nil, nil },
	}
	assert.Nil(t, cmd.Handle(mockContext))
}

func TestAgentsUpdateCommandUnsupportedVersion(t *testing.T) {
	os.MkdirAll(".ai", 0755)
	os.WriteFile(versionFilePath, []byte(`{"version":"v1.16","files":{}}`), 0644)
	defer os.RemoveAll(".ai")

	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Option("version").Return("v1.16").Once()
	mockContext.EXPECT().Error("Agent files are only available for Goravel v1.17 and above (got v1.16).").Once()

	cmd := &AgentsUpdateCommand{
		treeFetcher: func(branch string) ([]string, error) { return nil, nil },
		fetcher:     func(branch, path string) ([]byte, error) { return nil, nil },
	}
	assert.Nil(t, cmd.Handle(mockContext))
}

func TestAgentsUpdateCommandConflictDetection(t *testing.T) {
	var (
		mockContext *mocksconsole.Context
		cmd         *AgentsUpdateCommand
	)

	beforeEach := func() {
		mockContext = mocksconsole.NewContext(t)
		cmd = &AgentsUpdateCommand{}
	}

	cleanup := func() {
		os.RemoveAll(".ai")
		os.Remove("AGENTS.md")
	}

	originalContent := []byte("original content")
	storedChecksum := sha256sum(originalContent)

	upstreamContent := []byte("upstream changed content")
	userContent := []byte("user modified content")

	setupLocalVersionFile := func(checksum string) {
		vf := VersionFile{
			Version: "v1.17",
			Files:   map[string]string{"prompt/route.md": checksum},
		}
		data, _ := json.MarshalIndent(vf, "", "  ")
		os.MkdirAll(".ai/prompt", 0755)
		os.WriteFile(versionFilePath, data, 0644)
	}

	makeTreeFetcher := func() func(string) ([]string, error) {
		return func(branch string) ([]string, error) {
			return []string{"prompt/route.md"}, nil
		}
	}

	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "Conflict - user modified and upstream changed, no force",
			setup: func() {
				setupLocalVersionFile(storedChecksum)
				os.WriteFile(".ai/prompt/route.md", userContent, 0644)

				cmd.treeFetcher = makeTreeFetcher()
				cmd.fetcher = func(branch, path string) ([]byte, error) {
					return upstreamContent, nil
				}

				mockContext.EXPECT().Option("version").Return("v1.17").Once()
				mockContext.EXPECT().Option("file").Return("").Once()
				mockContext.EXPECT().OptionBool("force").Return(false).Once()
				mockContext.EXPECT().Warning("Conflict: prompt/route.md modified locally and changed upstream. Use --force to overwrite.").Once()
				mockContext.EXPECT().Info("0 updated, 0 skipped (user modified), 1 conflicts (use --force to overwrite), 0 already up to date.").Once()
			},
		},
		{
			name: "Conflict - user modified and upstream changed, force overwrites",
			setup: func() {
				setupLocalVersionFile(storedChecksum)
				os.WriteFile(".ai/prompt/route.md", userContent, 0644)

				cmd.treeFetcher = makeTreeFetcher()
				cmd.fetcher = func(branch, path string) ([]byte, error) {
					return upstreamContent, nil
				}

				mockContext.EXPECT().Option("version").Return("v1.17").Once()
				mockContext.EXPECT().Option("file").Return("").Once()
				mockContext.EXPECT().OptionBool("force").Return(true).Once()
				mockContext.EXPECT().Info("1 updated, 0 skipped (user modified), 0 conflicts (use --force to overwrite), 0 already up to date.").Once()
			},
		},
		{
			name: "Upstream changed, user did not modify - download",
			setup: func() {
				setupLocalVersionFile(storedChecksum)
				os.WriteFile(".ai/prompt/route.md", originalContent, 0644)

				cmd.treeFetcher = makeTreeFetcher()
				cmd.fetcher = func(branch, path string) ([]byte, error) {
					return upstreamContent, nil
				}

				mockContext.EXPECT().Option("version").Return("v1.17").Once()
				mockContext.EXPECT().Option("file").Return("").Once()
				mockContext.EXPECT().OptionBool("force").Return(false).Once()
				mockContext.EXPECT().Info("1 updated, 0 skipped (user modified), 0 conflicts (use --force to overwrite), 0 already up to date.").Once()
			},
		},
		{
			name: "User modified, upstream unchanged - skip",
			setup: func() {
				setupLocalVersionFile(storedChecksum)
				os.WriteFile(".ai/prompt/route.md", userContent, 0644)

				cmd.treeFetcher = makeTreeFetcher()
				cmd.fetcher = func(branch, path string) ([]byte, error) {
					return originalContent, nil
				}

				mockContext.EXPECT().Option("version").Return("v1.17").Once()
				mockContext.EXPECT().Option("file").Return("").Once()
				mockContext.EXPECT().OptionBool("force").Return(false).Once()
				mockContext.EXPECT().Info("0 updated, 1 skipped (user modified), 0 conflicts (use --force to overwrite), 0 already up to date.").Once()
			},
		},
		{
			name: "Already up to date",
			setup: func() {
				setupLocalVersionFile(storedChecksum)
				os.WriteFile(".ai/prompt/route.md", originalContent, 0644)

				cmd.treeFetcher = makeTreeFetcher()
				cmd.fetcher = func(branch, path string) ([]byte, error) {
					return originalContent, nil
				}

				mockContext.EXPECT().Option("version").Return("v1.17").Once()
				mockContext.EXPECT().Option("file").Return("").Once()
				mockContext.EXPECT().OptionBool("force").Return(false).Once()
				mockContext.EXPECT().Info("0 updated, 0 skipped (user modified), 0 conflicts (use --force to overwrite), 1 already up to date.").Once()
			},
		},
		{
			name: "New file in upstream - download",
			setup: func() {
				setupLocalVersionFile(storedChecksum)
				os.WriteFile(".ai/prompt/route.md", originalContent, 0644)

				cmd.treeFetcher = func(branch string) ([]string, error) {
					return []string{"prompt/route.md", "prompt/auth.md"}, nil
				}
				cmd.fetcher = func(branch, path string) ([]byte, error) {
					if path == "prompt/auth.md" {
						return []byte("auth content"), nil
					}
					return originalContent, nil
				}

				mockContext.EXPECT().Option("version").Return("v1.17").Once()
				mockContext.EXPECT().Option("file").Return("").Once()
				mockContext.EXPECT().OptionBool("force").Return(false).Once()
				mockContext.EXPECT().Info("1 updated, 0 skipped (user modified), 0 conflicts (use --force to overwrite), 1 already up to date.").Once()
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			cleanup()
			test.setup()
			defer cleanup()

			assert.Nil(t, cmd.Handle(mockContext))
		})
	}
}
