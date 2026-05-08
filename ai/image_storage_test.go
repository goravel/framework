package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goravel/framework/errors"
	mocksfilesystem "github.com/goravel/framework/mocks/filesystem"
)

func TestImageStorer_Store(t *testing.T) {
	originalStorageFacade := storageFacade
	t.Cleanup(func() {
		storageFacade = originalStorageFacade
	})

	storage := mocksfilesystem.NewStorage(t)
	storage.EXPECT().Put("avatar.png", "image").Return(nil).Once()
	storageFacade = storage

	storedPath, err := NewImageStorer().Store([]byte("image"), "avatar.png", "")

	assert.Equal(t, "avatar.png", storedPath)
	assert.NoError(t, err)
}

func TestImageStorer_StoreUsesDisk(t *testing.T) {
	originalStorageFacade := storageFacade
	t.Cleanup(func() {
		storageFacade = originalStorageFacade
	})

	storage := mocksfilesystem.NewStorage(t)
	driver := mocksfilesystem.NewDriver(t)
	storage.EXPECT().Disk("s3").Return(driver).Once()
	driver.EXPECT().Put("avatar.png", "image").Return(nil).Once()
	storageFacade = storage

	storedPath, err := NewImageStorer().Store([]byte("image"), "avatar.png", "s3")

	assert.Equal(t, "avatar.png", storedPath)
	assert.NoError(t, err)
}

func TestImageStorer_StoreAs(t *testing.T) {
	originalStorageFacade := storageFacade
	t.Cleanup(func() {
		storageFacade = originalStorageFacade
	})

	storage := mocksfilesystem.NewStorage(t)
	storage.EXPECT().Put("images/avatar.png", "image").Return(nil).Once()
	storageFacade = storage

	storedPath, err := NewImageStorer().StoreAs([]byte("image"), "images/avatar.png", "")

	assert.Equal(t, "images/avatar.png", storedPath)
	assert.NoError(t, err)
}

func TestImageStorer_StoreAsNormalizesWindowsSeparators(t *testing.T) {
	originalStorageFacade := storageFacade
	t.Cleanup(func() {
		storageFacade = originalStorageFacade
	})

	storage := mocksfilesystem.NewStorage(t)
	storage.EXPECT().Put("images/nested/avatar.png", "image").Return(nil).Once()
	storageFacade = storage

	storedPath, err := NewImageStorer().StoreAs([]byte("image"), `images\nested\avatar.png`, "")

	assert.Equal(t, "images/nested/avatar.png", storedPath)
	assert.NoError(t, err)
}

func TestImageStorer_StoreAsRequiresFileName(t *testing.T) {
	tests := []struct {
		name       string
		targetPath string
		expectErr  error
	}{
		{
			name:       "empty path",
			targetPath: "",
			expectErr:  errors.AIImageStorePathRequired,
		},
		{
			name:       "unix directory path",
			targetPath: "images/",
			expectErr:  errors.AIImageNameRequired,
		},
		{
			name:       "windows directory path",
			targetPath: `images\nested\`,
			expectErr:  errors.AIImageNameRequired,
		},
		{
			name:       "parent segment",
			targetPath: "../avatar.png",
			expectErr:  errors.AIImageStorePathInvalid,
		},
		{
			name:       "absolute path",
			targetPath: "/tmp/avatar.png",
			expectErr:  errors.AIImageStorePathInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storedPath, err := NewImageStorer().StoreAs([]byte("image"), tt.targetPath, "")

			assert.Equal(t, "", storedPath)
			assert.Equal(t, tt.expectErr, err)
		})
	}
}

func TestImageStorer_StoreReturnsErrorWithoutStorageFacade(t *testing.T) {
	originalStorageFacade := storageFacade
	t.Cleanup(func() {
		storageFacade = originalStorageFacade
	})
	storageFacade = nil

	storedPath, err := NewImageStorer().Store([]byte("image"), "avatar.png", "")

	assert.Equal(t, "", storedPath)
	assert.Equal(t, errors.StorageFacadeNotSet, err)
}

func TestImageResponse_Store(t *testing.T) {
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
				storage.EXPECT().Put("fixed-name.png", "image").Return(nil).Once()
			},
			expectPath: "fixed-name.png",
		},
		{
			name: "explicit disk",
			disk: []string{"s3"},
			setup: func(storage *mocksfilesystem.Storage, driver *mocksfilesystem.Driver) {
				storage.EXPECT().Disk("s3").Return(driver).Once()
				driver.EXPECT().Put("fixed-name.png", "image").Return(nil).Once()
			},
			expectPath: "fixed-name.png",
		},
		{
			name:      "too many disks",
			disk:      []string{"s3", "local"},
			expectErr: errors.AIImageStoreTooManyPaths,
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

			response := &imageResponse{content: []byte("image"), mimeType: "image/png", name: "fixed-name.png", storer: NewImageStorer()}

			storedPath, err := response.Store(tt.disk...)

			assert.Equal(t, tt.expectPath, storedPath)
			assert.Equal(t, tt.expectErr, err)
		})
	}
}

func TestImageResponse_StoreAs(t *testing.T) {
	originalStorageFacade := storageFacade
	t.Cleanup(func() {
		storageFacade = originalStorageFacade
	})

	storage := mocksfilesystem.NewStorage(t)
	storage.EXPECT().Put("images/avatar.png", "image").Return(nil).Once()
	storageFacade = storage

	response := &imageResponse{content: []byte("image"), mimeType: "image/png", storer: NewImageStorer()}

	storedPath, err := response.StoreAs("images/avatar.png")

	assert.Equal(t, "images/avatar.png", storedPath)
	assert.NoError(t, err)
}

func TestImageResponse_ContentClonesBytes(t *testing.T) {
	response := &imageResponse{content: []byte("image")}

	content, err := response.Content()
	require.NoError(t, err)
	content[0] = 'I'

	assert.Equal(t, []byte("image"), response.content)
}
