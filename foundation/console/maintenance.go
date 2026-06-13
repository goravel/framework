package console

import (
	"strings"

	"github.com/goravel/framework/contracts/cache"
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/errors"
)

const (
	maintenanceCacheKey    = "framework:maintenance"
	maintenanceDriverCache = "cache"
	maintenanceDriverFile  = "file"
	maintenanceFilePath    = "framework/maintenance.json"
)

type MaintenanceMode struct {
	cache   cache.Cache
	config  config.Config
	storage filesystem.Storage
}

func NewMaintenanceMode(config config.Config, cache cache.Cache, storage filesystem.Storage) *MaintenanceMode {
	return &MaintenanceMode{
		cache:   cache,
		config:  config,
		storage: storage,
	}
}

func (r *MaintenanceMode) Get() ([]byte, bool, error) {
	driver := r.driver()
	switch driver {
	case maintenanceDriverFile:
		if r.storage == nil {
			return nil, false, errors.StorageFacadeNotSet
		}
		if !r.storage.Exists(maintenanceFilePath) {
			return nil, false, nil
		}

		content, err := r.storage.GetBytes(maintenanceFilePath)
		return content, true, err
	case maintenanceDriverCache:
		cacheStore, err := r.cacheDriver()
		if err != nil {
			return nil, false, err
		}
		if !cacheStore.Has(maintenanceCacheKey) {
			return nil, false, nil
		}

		return []byte(cacheStore.GetString(maintenanceCacheKey)), true, nil
	default:
		return nil, false, errors.MaintenanceDriverNotSupported.Args(driver)
	}
}

func (r *MaintenanceMode) Put(content string) error {
	driver := r.driver()
	switch driver {
	case maintenanceDriverFile:
		if r.storage == nil {
			return errors.StorageFacadeNotSet
		}

		return r.storage.Put(maintenanceFilePath, content)
	case maintenanceDriverCache:
		cacheStore, err := r.cacheDriver()
		if err != nil {
			return err
		}
		if !cacheStore.Forever(maintenanceCacheKey, content) {
			return errors.MaintenanceCachePutFailed
		}

		return nil
	default:
		return errors.MaintenanceDriverNotSupported.Args(driver)
	}
}

func (r *MaintenanceMode) Delete() (bool, error) {
	driver := r.driver()
	switch driver {
	case maintenanceDriverFile:
		if r.storage == nil {
			return false, errors.StorageFacadeNotSet
		}
		if !r.storage.Exists(maintenanceFilePath) {
			return false, nil
		}

		return true, r.storage.Delete(maintenanceFilePath)
	case maintenanceDriverCache:
		cacheStore, err := r.cacheDriver()
		if err != nil {
			return false, err
		}
		if !cacheStore.Has(maintenanceCacheKey) {
			return false, nil
		}
		if !cacheStore.Forget(maintenanceCacheKey) {
			return true, errors.MaintenanceCacheDeleteFailed
		}

		return true, nil
	default:
		return false, errors.MaintenanceDriverNotSupported.Args(driver)
	}
}

func (r *MaintenanceMode) cacheDriver() (cache.Driver, error) {
	if r.cache == nil {
		return nil, errors.CacheFacadeNotSet
	}

	store := r.store()
	if store == "" {
		return r.cache, nil
	}

	driver := r.cache.Store(store)
	if driver == nil {
		return nil, errors.MaintenanceCacheStoreNotFound.Args(store)
	}

	return driver, nil
}

func (r *MaintenanceMode) driver() string {
	if r.config == nil {
		return maintenanceDriverFile
	}

	driver := strings.ToLower(r.config.GetString("app.maintenance.driver", maintenanceDriverFile))
	if driver == "" {
		return maintenanceDriverFile
	}

	return driver
}

func (r *MaintenanceMode) store() string {
	if r.config == nil {
		return ""
	}

	return r.config.GetString("app.maintenance.store")
}
