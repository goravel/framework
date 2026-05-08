package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goravel/framework/errors"
	mocksfilesystem "github.com/goravel/framework/mocks/filesystem"
)

func TestAudioStorer_Store(t *testing.T) {
	originalStorageFacade := storageFacade
	t.Cleanup(func() {
		storageFacade = originalStorageFacade
	})

	storage := mocksfilesystem.NewStorage(t)
	storage.EXPECT().Put("voice.mp3", "audio").Return(nil).Once()
	storageFacade = storage

	storedPath, err := audioStorer{}.Store([]byte("audio"), "voice.mp3", "")

	assert.Equal(t, "voice.mp3", storedPath)
	assert.NoError(t, err)
}

func TestAudioStorer_StoreUsesDisk(t *testing.T) {
	originalStorageFacade := storageFacade
	t.Cleanup(func() {
		storageFacade = originalStorageFacade
	})

	storage := mocksfilesystem.NewStorage(t)
	driver := mocksfilesystem.NewDriver(t)
	storage.EXPECT().Disk("s3").Return(driver).Once()
	driver.EXPECT().Put("voice.mp3", "audio").Return(nil).Once()
	storageFacade = storage

	storedPath, err := audioStorer{}.Store([]byte("audio"), "voice.mp3", "s3")

	assert.Equal(t, "voice.mp3", storedPath)
	assert.NoError(t, err)
}

func TestAudioStorer_StoreAs(t *testing.T) {
	originalStorageFacade := storageFacade
	t.Cleanup(func() {
		storageFacade = originalStorageFacade
	})

	storage := mocksfilesystem.NewStorage(t)
	storage.EXPECT().Put("audio/voice.mp3", "audio").Return(nil).Once()
	storageFacade = storage

	storedPath, err := audioStorer{}.StoreAs([]byte("audio"), "audio/voice.mp3", "")

	assert.Equal(t, "audio/voice.mp3", storedPath)
	assert.NoError(t, err)
}

func TestAudioStorer_StoreAsNormalizesWindowsSeparators(t *testing.T) {
	originalStorageFacade := storageFacade
	t.Cleanup(func() {
		storageFacade = originalStorageFacade
	})

	storage := mocksfilesystem.NewStorage(t)
	storage.EXPECT().Put("audio/nested/voice.mp3", "audio").Return(nil).Once()
	storageFacade = storage

	storedPath, err := audioStorer{}.StoreAs([]byte("audio"), `audio\nested\voice.mp3`, "")

	assert.Equal(t, "audio/nested/voice.mp3", storedPath)
	assert.NoError(t, err)
}

func TestAudioStorer_StoreAsRejectsInvalidPaths(t *testing.T) {
	tests := []struct {
		name       string
		targetPath string
		expectErr  error
	}{
		{
			name:       "empty path",
			targetPath: "",
			expectErr:  errors.AIAudioStorePathRequired,
		},
		{
			name:       "unix directory path",
			targetPath: "audio/",
			expectErr:  errors.AIAudioNameRequired,
		},
		{
			name:       "windows directory path",
			targetPath: `audio\nested\`,
			expectErr:  errors.AIAudioNameRequired,
		},
		{
			name:       "parent segment",
			targetPath: "../voice.mp3",
			expectErr:  errors.AIAudioStorePathInvalid,
		},
		{
			name:       "absolute path",
			targetPath: "/tmp/voice.mp3",
			expectErr:  errors.AIAudioStorePathInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storedPath, err := audioStorer{}.StoreAs([]byte("audio"), tt.targetPath, "")

			assert.Equal(t, "", storedPath)
			assert.Equal(t, tt.expectErr, err)
		})
	}
}

func TestAudioStorer_StoreReturnsErrorWithoutStorageFacade(t *testing.T) {
	originalStorageFacade := storageFacade
	t.Cleanup(func() {
		storageFacade = originalStorageFacade
	})
	storageFacade = nil

	storedPath, err := audioStorer{}.Store([]byte("audio"), "voice.mp3", "")

	assert.Equal(t, "", storedPath)
	assert.Equal(t, errors.StorageFacadeNotSet, err)
}

func TestAudioResponse_Store(t *testing.T) {
	tests := []struct {
		name       string
		disk       []string
		setup      func(*mocksfilesystem.Storage, *mocksfilesystem.Driver)
		expectPath string
		expectErr  error
	}{
		{
			name: "default disk",
			setup: func(storage *mocksfilesystem.Storage, _ *mocksfilesystem.Driver) {
				storage.EXPECT().Put("fixed-name.mp3", "audio").Return(nil).Once()
			},
			expectPath: "fixed-name.mp3",
		},
		{
			name: "explicit disk",
			disk: []string{"s3"},
			setup: func(storage *mocksfilesystem.Storage, driver *mocksfilesystem.Driver) {
				storage.EXPECT().Disk("s3").Return(driver).Once()
				driver.EXPECT().Put("fixed-name.mp3", "audio").Return(nil).Once()
			},
			expectPath: "fixed-name.mp3",
		},
		{
			name:      "too many disks",
			disk:      []string{"s3", "local"},
			expectErr: errors.AIAudioStoreTooManyPaths,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalStorageFacade := storageFacade
			t.Cleanup(func() {
				storageFacade = originalStorageFacade
			})

			storage := mocksfilesystem.NewStorage(t)
			driver := mocksfilesystem.NewDriver(t)
			if tt.setup != nil {
				tt.setup(storage, driver)
			}
			storageFacade = storage

			response := &audioResponse{content: []byte("audio"), mimeType: "audio/mpeg", name: "fixed-name.mp3", storer: audioStorer{}}

			storedPath, err := response.Store(tt.disk...)

			assert.Equal(t, tt.expectPath, storedPath)
			assert.Equal(t, tt.expectErr, err)
		})
	}
}

func TestAudioResponse_StoreAs(t *testing.T) {
	originalStorageFacade := storageFacade
	t.Cleanup(func() {
		storageFacade = originalStorageFacade
	})

	storage := mocksfilesystem.NewStorage(t)
	storage.EXPECT().Put("audio/voice.mp3", "audio").Return(nil).Once()
	storageFacade = storage

	response := &audioResponse{content: []byte("audio"), mimeType: "audio/mpeg", storer: audioStorer{}}

	storedPath, err := response.StoreAs("audio/voice.mp3")

	assert.Equal(t, "audio/voice.mp3", storedPath)
	assert.NoError(t, err)
}

func TestAudioResponse_ContentClonesBytes(t *testing.T) {
	response := &audioResponse{content: []byte("audio")}

	content, err := response.Content()
	require.NoError(t, err)
	content[0] = 'A'

	assert.Equal(t, []byte("audio"), response.content)
}

func TestAudioResponse_StorageNameUsesMimeType(t *testing.T) {
	tests := []struct {
		name      string
		mimeType  string
		suffix    string
	}{
		{name: "mp3", mimeType: "audio/mpeg", suffix: ".mp3"},
		{name: "wav", mimeType: "audio/wav", suffix: ".wav"},
		{name: "flac", mimeType: "audio/flac", suffix: ".flac"},
		{name: "aac", mimeType: "audio/aac", suffix: ".aac"},
		{name: "opus", mimeType: "audio/opus", suffix: ".opus"},
		{name: "pcm", mimeType: "audio/pcm", suffix: ".pcm"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := &audioResponse{mimeType: tt.mimeType}
			assert.Contains(t, response.storageName(), tt.suffix)
		})
	}
}
