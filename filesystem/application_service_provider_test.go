package filesystem

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/binding"
	contractsfilesystem "github.com/goravel/framework/contracts/filesystem"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksfilesystem "github.com/goravel/framework/mocks/filesystem"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
)

func TestNewStorage(t *testing.T) {
	t.Run("returns error when default disk is not configured", func(t *testing.T) {
		config := mocksconfig.NewConfig(t)
		config.EXPECT().GetString("filesystems.default").Return("").Once()

		storage, err := NewStorage(config)

		assert.Nil(t, storage)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), errors.FilesystemDefaultDiskNotSet.Error())
	})

	t.Run("creates storage and caches loaded disk", func(t *testing.T) {
		config := mocksconfig.NewConfig(t)
		config.EXPECT().GetString("filesystems.default").Return("local").Once()
		config.EXPECT().GetString("filesystems.disks.local.driver").Return("local").Once()
		config.EXPECT().GetString("filesystems.disks.local.root").Return(t.TempDir()).Once()
		config.EXPECT().GetString("filesystems.disks.local.url").Return("").Once()

		config.EXPECT().GetString("filesystems.disks.backup.driver").Return("local").Once()
		config.EXPECT().GetString("filesystems.disks.backup.root").Return(t.TempDir()).Once()
		config.EXPECT().GetString("filesystems.disks.backup.url").Return("").Once()

		storage, err := NewStorage(config)

		assert.NoError(t, err)
		assert.NotNil(t, storage)
		assert.Same(t, storage.Driver, storage.Disk("local"))

		first := storage.Disk("backup")
		second := storage.Disk("backup")
		assert.Same(t, first, second)
	})
}

func TestNewDriver(t *testing.T) {
	t.Run("returns local driver", func(t *testing.T) {
		config := mocksconfig.NewConfig(t)
		config.EXPECT().GetString("filesystems.disks.local.driver").Return("local").Once()
		config.EXPECT().GetString("filesystems.disks.local.root").Return(t.TempDir()).Once()
		config.EXPECT().GetString("filesystems.disks.local.url").Return("").Once()

		driver, err := NewDriver(config, "local")

		assert.NoError(t, err)
		assert.IsType(t, &Local{}, driver)
	})

	t.Run("returns custom driver instance", func(t *testing.T) {
		config := mocksconfig.NewConfig(t)
		custom := mocksfilesystem.NewDriver(t)
		config.EXPECT().GetString("filesystems.disks.custom.driver").Return("custom").Once()
		config.EXPECT().Get("filesystems.disks.custom.via").Return(custom).Once()

		driver, err := NewDriver(config, "custom")

		assert.NoError(t, err)
		assert.Same(t, custom, driver)
	})

	t.Run("returns custom driver from callback", func(t *testing.T) {
		config := mocksconfig.NewConfig(t)
		custom := mocksfilesystem.NewDriver(t)
		config.EXPECT().GetString("filesystems.disks.custom.driver").Return("custom").Once()
		config.EXPECT().Get("filesystems.disks.custom.via").Return(func() (contractsfilesystem.Driver, error) {
			return custom, nil
		}).Once()

		driver, err := NewDriver(config, "custom")

		assert.NoError(t, err)
		assert.Same(t, custom, driver)
	})

	t.Run("returns error for invalid custom driver", func(t *testing.T) {
		config := mocksconfig.NewConfig(t)
		config.EXPECT().GetString("filesystems.disks.custom.driver").Return("custom").Once()
		config.EXPECT().Get("filesystems.disks.custom.via").Return("invalid").Once()

		driver, err := NewDriver(config, "custom")

		assert.Nil(t, driver)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "via must be implement filesystem.Driver or func() (filesystem.Driver, error)")
	})

	t.Run("returns error for unsupported driver", func(t *testing.T) {
		config := mocksconfig.NewConfig(t)
		config.EXPECT().GetString("filesystems.disks.s3.driver").Return("s3").Once()

		driver, err := NewDriver(config, "s3")

		assert.Nil(t, driver)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "only support local, custom")
	})
}

func TestStorageDisk_PanicOnInvalidDisk(t *testing.T) {
	config := mocksconfig.NewConfig(t)
	config.EXPECT().GetString("filesystems.default").Return("local").Once()
	config.EXPECT().GetString("filesystems.disks.local.driver").Return("local").Once()
	config.EXPECT().GetString("filesystems.disks.local.root").Return(t.TempDir()).Once()
	config.EXPECT().GetString("filesystems.disks.local.url").Return("").Once()
	config.EXPECT().GetString("filesystems.disks.invalid.driver").Return("unsupported").Once()

	storage, err := NewStorage(config)
	assert.NoError(t, err)

	assert.Panics(t, func() {
		storage.Disk("invalid")
	})
}

func TestServiceProvider(t *testing.T) {
	provider := &ServiceProvider{}

	t.Run("relationship", func(t *testing.T) {
		relationship := provider.Relationship()

		assert.Equal(t, []string{binding.Storage}, relationship.Bindings)
		assert.Equal(t, binding.Bindings[binding.Storage].Dependencies, relationship.Dependencies)
		assert.Empty(t, relationship.ProvideFor)
	})

	t.Run("register returns error when config facade not set", func(t *testing.T) {
		app := mocksfoundation.NewApplication(t)
		app.EXPECT().Singleton(binding.Storage, mock.AnythingOfType("func(foundation.Application) (interface {}, error)")).Run(func(_ any, callback func(contractsfoundation.Application) (any, error)) {
			callbackApp := mocksfoundation.NewApplication(t)
			callbackApp.EXPECT().MakeConfig().Return(nil).Once()

			instance, err := callback(callbackApp)

			assert.Nil(t, instance)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), errors.ConfigFacadeNotSet.Error())
		}).Once()

		provider.Register(app)
	})

	t.Run("register creates storage singleton", func(t *testing.T) {
		app := mocksfoundation.NewApplication(t)
		app.EXPECT().Singleton(binding.Storage, mock.AnythingOfType("func(foundation.Application) (interface {}, error)")).Run(func(_ any, callback func(contractsfoundation.Application) (any, error)) {
			callbackApp := mocksfoundation.NewApplication(t)
			config := mocksconfig.NewConfig(t)
			callbackApp.EXPECT().MakeConfig().Return(config).Once()
			config.EXPECT().GetString("filesystems.default").Return("local").Once()
			config.EXPECT().GetString("filesystems.disks.local.driver").Return("local").Once()
			config.EXPECT().GetString("filesystems.disks.local.root").Return(t.TempDir()).Once()
			config.EXPECT().GetString("filesystems.disks.local.url").Return("").Once()

			instance, err := callback(callbackApp)

			assert.NoError(t, err)
			assert.IsType(t, &Storage{}, instance)
		}).Once()

		provider.Register(app)
	})

	t.Run("boot sets facades", func(t *testing.T) {
		originConfigFacade := ConfigFacade
		originStorageFacade := StorageFacade
		t.Cleanup(func() {
			ConfigFacade = originConfigFacade
			StorageFacade = originStorageFacade
		})

		app := mocksfoundation.NewApplication(t)
		config := mocksconfig.NewConfig(t)
		storage := mocksfilesystem.NewStorage(t)
		app.EXPECT().MakeConfig().Return(config).Once()
		app.EXPECT().MakeStorage().Return(storage).Once()

		provider.Boot(app)

		assert.Same(t, config, ConfigFacade)
		assert.Same(t, storage, StorageFacade)
	})
}
