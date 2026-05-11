package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/errors"
)

func TestNormalizeStoreTargetPath(t *testing.T) {
	pathErrors := storePathErrors{
		pathRequired: errors.AIAudioStorePathRequired,
		nameRequired: errors.AIAudioNameRequired,
		pathInvalid:  errors.AIAudioStorePathInvalid,
	}

	tests := []struct {
		name            string
		targetPath      string
		expectName      string
		expectDirectory string
		expectErr       error
	}{
		{
			name:            "normalizes windows separators",
			targetPath:      `audio\nested\voice.mp3`,
			expectName:      "voice.mp3",
			expectDirectory: "audio/nested",
		},
		{
			name:      "rejects parent segment",
			targetPath: "../voice.mp3",
			expectErr: errors.AIAudioStorePathInvalid,
		},
		{
			name:      "rejects absolute unix path",
			targetPath: "/tmp/voice.mp3",
			expectErr: errors.AIAudioStorePathInvalid,
		},
		{
			name:      "rejects windows volume path",
			targetPath: `C:\tmp\voice.mp3`,
			expectErr: errors.AIAudioStorePathInvalid,
		},
		{
			name:      "requires file name",
			targetPath: "audio/",
			expectErr: errors.AIAudioNameRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, directory, err := normalizeStoreTargetPath(tt.targetPath, pathErrors)

			assert.Equal(t, tt.expectName, name)
			assert.Equal(t, tt.expectDirectory, directory)
			assert.Equal(t, tt.expectErr, err)
		})
	}
}

func TestIsAbsoluteStoragePath(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		expect bool
	}{
		{name: "relative path", path: "audio/voice.mp3", expect: false},
		{name: "absolute unix path", path: "/tmp/voice.mp3", expect: true},
		{name: "windows volume path", path: "C:/tmp/voice.mp3", expect: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, isAbsoluteStoragePath(tt.path))
		})
	}
}
