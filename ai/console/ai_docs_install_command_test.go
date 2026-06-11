package console

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	mocksconsole "github.com/goravel/framework/mocks/console"
)

func TestAiDocsInstallCommand(t *testing.T) {
	var (
		mockContext    *mocksconsole.Context
		installCommand *AiDocsInstallCommand
	)

	beforeEach := func() {
		mockContext = mocksconsole.NewContext(t)
		installCommand = &AiDocsInstallCommand{
			versionDetector: func() (string, error) { return "v1.17", nil },
		}
	}

	cleanup := func() {
		assert.Nil(t, os.RemoveAll(".ai"))
		assert.Nil(t, os.RemoveAll("AGENTS.md"))
	}

	manifest := []ManifestEntry{
		{Facade: "", Path: "AGENTS.md", Default: true},
		{Facade: "Route", Path: "prompt/route.md", Default: true},
		{Facade: "Auth", Path: "prompt/auth.md", Default: false},
	}

	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "Happy path - installs defaults when no facades given",
			setup: func() {
				installCommand.manifestFetcher = func(branch string) ([]ManifestEntry, error) {
					return manifest, nil
				}
				installCommand.fetcher = func(branch, path string) ([]byte, error) {
					return []byte("# " + path), nil
				}

				mockContext.EXPECT().OptionBool("force").Return(true).Once()
				mockContext.EXPECT().Arguments().Return([]string{}).Once()
				mockContext.EXPECT().OptionBool("all").Return(false).Once()
				mockContext.EXPECT().Info("Installed 2 file(s) for version v1.17.").Once()
			},
		},
		{
			name: "Happy path - installs all when --all flag set",
			setup: func() {
				installCommand.manifestFetcher = func(branch string) ([]ManifestEntry, error) {
					return manifest, nil
				}
				installCommand.fetcher = func(branch, path string) ([]byte, error) {
					return []byte("# " + path), nil
				}

				mockContext.EXPECT().OptionBool("force").Return(true).Once()
				mockContext.EXPECT().Arguments().Return([]string{}).Once()
				mockContext.EXPECT().OptionBool("all").Return(true).Once()
				mockContext.EXPECT().Info("Installed 3 file(s) for version v1.17.").Once()
			},
		},
		{
			name: "Happy path - installs specific facade by name",
			setup: func() {
				installCommand.manifestFetcher = func(branch string) ([]ManifestEntry, error) {
					return manifest, nil
				}
				installCommand.fetcher = func(branch, path string) ([]byte, error) {
					return []byte("# " + path), nil
				}

				mockContext.EXPECT().OptionBool("force").Return(true).Once()
				mockContext.EXPECT().Arguments().Return([]string{"Auth"}).Once()
				mockContext.EXPECT().OptionBool("all").Return(false).Once()
				mockContext.EXPECT().Info("Installed 1 file(s) for version v1.17.").Once()
			},
		},
		{
			name: "Happy path - falls back to master when version branch has no manifest",
			setup: func() {
				installCommand.versionDetector = func() (string, error) { return "v1.99", nil }
				installCommand.manifestFetcher = func(branch string) ([]ManifestEntry, error) {
					if branch == docsFallbackBranch {
						return manifest, nil
					}
					return nil, nil
				}
				installCommand.fetcher = func(branch, path string) ([]byte, error) {
					return []byte("# " + path), nil
				}

				mockContext.EXPECT().OptionBool("force").Return(true).Once()
				mockContext.EXPECT().Arguments().Return([]string{}).Once()
				mockContext.EXPECT().OptionBool("all").Return(false).Once()
				mockContext.EXPECT().Info("Installed 2 file(s) for version v1.99.").Once()
			},
		},
		{
			name: "Sad path - no manifest on branch or master",
			setup: func() {
				installCommand.versionDetector = func() (string, error) { return "v9.99", nil }
				installCommand.manifestFetcher = func(branch string) ([]ManifestEntry, error) {
					return nil, nil
				}

				mockContext.EXPECT().Error("No AI docs found for version v9.99. Check https://github.com/goravel/docs").Once()
			},
		},
		{
			name: "Sad path - manifest fetch error",
			setup: func() {
				installCommand.manifestFetcher = func(branch string) ([]ManifestEntry, error) {
					return nil, errors.New("network error")
				}

				mockContext.EXPECT().Error("network error").Once()
			},
		},
		{
			name: "Sad path - facade not found in manifest",
			setup: func() {
				installCommand.manifestFetcher = func(branch string) ([]ManifestEntry, error) {
					return manifest, nil
				}

				mockContext.EXPECT().OptionBool("force").Return(true).Once()
				mockContext.EXPECT().Arguments().Return([]string{"Nonexistent"}).Once()
				mockContext.EXPECT().OptionBool("all").Return(false).Once()
				mockContext.EXPECT().Error("No AI docs found for facade(s): Nonexistent").Once()
			},
		},
		{
			name: "Sad path - unsupported version",
			setup: func() {
				installCommand.versionDetector = func() (string, error) { return "v1.16", nil }
				mockContext.EXPECT().Error("AI docs are only available for Goravel v1.17 and above (got v1.16)").Once()
			},
		},
		{
			name: "Sad path - existing install, user cancels",
			setup: func() {
				installCommand.manifestFetcher = func(branch string) ([]ManifestEntry, error) {
					return manifest, nil
				}

				assert.Nil(t, os.MkdirAll(".ai", 0755))
				assert.Nil(t, os.WriteFile(versionFilePath, []byte(`{"version":"v1.17","files":{}}`), 0644))

				mockContext.EXPECT().OptionBool("force").Return(false).Once()
				mockContext.EXPECT().Confirm("AI docs are already installed. Overwrite?").Return(false).Once()
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
