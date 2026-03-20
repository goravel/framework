package console

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/errors" // Adjust import path as needed
	mocksconsole "github.com/goravel/framework/mocks/console"
)

func TestAiDocsUpdateCommandMissingVersionFile(t *testing.T) {
	assert.Nil(t, os.RemoveAll(".ai"))

	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Error(errors.AiDocsNotInstalled.Error()).Once()

	cmd := &AiDocsUpdateCommand{
		manifestFetcher: func(branch string) ([]ManifestEntry, error) { return nil, nil },
		fetcher:         func(branch, path string) ([]byte, error) { return nil, nil },
	}
	assert.Nil(t, cmd.Handle(mockContext))
}

func TestAiDocsUpdateCommandNoDocsFound(t *testing.T) {
	assert.Nil(t, os.MkdirAll(".ai", 0755))
	assert.Nil(t, os.WriteFile(versionFilePath, []byte(`{"version":"v1.16","files":{}}`), 0644))
	defer func() { assert.Nil(t, os.RemoveAll(".ai")) }()

	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Error(errors.AiDocsManifestFailed.Error()).Once()

	cmd := &AiDocsUpdateCommand{
		manifestFetcher: func(branch string) ([]ManifestEntry, error) { return nil, nil },
		fetcher:         func(branch, path string) ([]byte, error) { return nil, nil },
	}
	assert.Nil(t, cmd.Handle(mockContext))
}

func TestAiDocsUpdateCommandConflictDetection(t *testing.T) {
	var (
		mockContext *mocksconsole.Context
		cmd         *AiDocsUpdateCommand
	)

	beforeEach := func() {
		mockContext = mocksconsole.NewContext(t)
		cmd = &AiDocsUpdateCommand{}
	}

	cleanup := func() {
		assert.Nil(t, os.RemoveAll(".ai"))
		assert.Nil(t, os.RemoveAll("AGENTS.md"))
	}

	originalContent := []byte("original content")
	storedChecksum := sha256sum(originalContent)

	upstreamContent := []byte("upstream changed content")
	userContent := []byte("user modified content")

	routeEntry := ManifestEntry{Facade: "Route", Path: "prompt/route.md", Default: true}

	setupLocalVersionFile := func(checksum string) {
		vf := VersionFile{
			Version: "v1.17",
			Files:   map[string]string{"prompt/route.md": checksum},
		}
		data, err := json.MarshalIndent(vf, "", "  ")
		assert.Nil(t, err)

		assert.Nil(t, os.MkdirAll(".ai/prompt", 0755))
		assert.Nil(t, os.WriteFile(versionFilePath, data, 0644))
	}

	makeManifestFetcher := func() func(string) ([]ManifestEntry, error) {
		return func(branch string) ([]ManifestEntry, error) {
			return []ManifestEntry{routeEntry}, nil
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
				assert.Nil(t, os.WriteFile(".ai/prompt/route.md", userContent, 0644))

				cmd.manifestFetcher = makeManifestFetcher()
				cmd.fetcher = func(branch, path string) ([]byte, error) {
					return upstreamContent, nil
				}

				mockContext.EXPECT().Arguments().Return([]string{}).Once()
				mockContext.EXPECT().OptionBool("all").Return(false).Once()
				mockContext.EXPECT().OptionBool("force").Return(false).Once()
				mockContext.EXPECT().Warning("Conflict: prompt/route.md modified locally and changed upstream. Use --force to overwrite.").Once()
				mockContext.EXPECT().Info("0 updated, 0 skipped (user modified), 1 conflicts (use --force), 0 up to date.").Once()
			},
		},
		{
			name: "Conflict - user modified and upstream changed, force overwrites",
			setup: func() {
				setupLocalVersionFile(storedChecksum)
				assert.Nil(t, os.WriteFile(".ai/prompt/route.md", userContent, 0644))

				cmd.manifestFetcher = makeManifestFetcher()
				cmd.fetcher = func(branch, path string) ([]byte, error) {
					return upstreamContent, nil
				}

				mockContext.EXPECT().Arguments().Return([]string{}).Once()
				mockContext.EXPECT().OptionBool("all").Return(false).Once()
				mockContext.EXPECT().OptionBool("force").Return(true).Once()
				mockContext.EXPECT().Info("1 updated, 0 skipped (user modified), 0 conflicts (use --force), 0 up to date.").Once()
			},
		},
		{
			name: "Upstream changed, user did not modify - download",
			setup: func() {
				setupLocalVersionFile(storedChecksum)
				assert.Nil(t, os.WriteFile(".ai/prompt/route.md", originalContent, 0644))

				cmd.manifestFetcher = makeManifestFetcher()
				cmd.fetcher = func(branch, path string) ([]byte, error) {
					return upstreamContent, nil
				}

				mockContext.EXPECT().Arguments().Return([]string{}).Once()
				mockContext.EXPECT().OptionBool("all").Return(false).Once()
				mockContext.EXPECT().OptionBool("force").Return(false).Once()
				mockContext.EXPECT().Info("1 updated, 0 skipped (user modified), 0 conflicts (use --force), 0 up to date.").Once()
			},
		},
		{
			name: "User modified, upstream unchanged - skip",
			setup: func() {
				setupLocalVersionFile(storedChecksum)
				assert.Nil(t, os.WriteFile(".ai/prompt/route.md", userContent, 0644))

				cmd.manifestFetcher = makeManifestFetcher()
				cmd.fetcher = func(branch, path string) ([]byte, error) {
					return originalContent, nil
				}

				mockContext.EXPECT().Arguments().Return([]string{}).Once()
				mockContext.EXPECT().OptionBool("all").Return(false).Once()
				mockContext.EXPECT().OptionBool("force").Return(false).Once()
				mockContext.EXPECT().Info("0 updated, 1 skipped (user modified), 0 conflicts (use --force), 0 up to date.").Once()
			},
		},
		{
			name: "Already up to date",
			setup: func() {
				setupLocalVersionFile(storedChecksum)
				assert.Nil(t, os.WriteFile(".ai/prompt/route.md", originalContent, 0644))

				cmd.manifestFetcher = makeManifestFetcher()
				cmd.fetcher = func(branch, path string) ([]byte, error) {
					return originalContent, nil
				}

				mockContext.EXPECT().Arguments().Return([]string{}).Once()
				mockContext.EXPECT().OptionBool("all").Return(false).Once()
				mockContext.EXPECT().OptionBool("force").Return(false).Once()
				mockContext.EXPECT().Info("0 updated, 0 skipped (user modified), 0 conflicts (use --force), 1 up to date.").Once()
			},
		},
		{
			name: "New file in manifest with --all - download",
			setup: func() {
				setupLocalVersionFile(storedChecksum)
				assert.Nil(t, os.WriteFile(".ai/prompt/route.md", originalContent, 0644))

				cmd.manifestFetcher = func(branch string) ([]ManifestEntry, error) {
					return []ManifestEntry{
						routeEntry,
						{Facade: "Auth", Path: "prompt/auth.md", Default: false},
					}, nil
				}
				cmd.fetcher = func(branch, path string) ([]byte, error) {
					if path == "prompt/auth.md" {
						return []byte("auth content"), nil
					}
					return originalContent, nil
				}

				mockContext.EXPECT().Arguments().Return([]string{}).Once()
				mockContext.EXPECT().OptionBool("all").Return(true).Once()
				mockContext.EXPECT().OptionBool("force").Return(false).Once()
				mockContext.EXPECT().Info("1 updated, 0 skipped (user modified), 0 conflicts (use --force), 1 up to date.").Once()
			},
		},
		{
			name: "Specific facade via argument",
			setup: func() {
				setupLocalVersionFile(storedChecksum)
				assert.Nil(t, os.WriteFile(".ai/prompt/route.md", originalContent, 0644))

				cmd.manifestFetcher = func(branch string) ([]ManifestEntry, error) {
					return []ManifestEntry{
						routeEntry,
						{Facade: "Auth", Path: "prompt/auth.md", Default: false},
					}, nil
				}
				cmd.fetcher = func(branch, path string) ([]byte, error) {
					return upstreamContent, nil
				}

				mockContext.EXPECT().Arguments().Return([]string{"Route"}).Once()
				mockContext.EXPECT().OptionBool("force").Return(false).Once()
				mockContext.EXPECT().Info("1 updated, 0 skipped (user modified), 0 conflicts (use --force), 0 up to date.").Once()
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
