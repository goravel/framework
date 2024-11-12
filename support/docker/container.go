package docker

import (
	"bytes"
	"github.com/goravel/framework/contracts/testing"
	"github.com/goravel/framework/foundation/json"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/file"
	"io"
	"os"
	"path/filepath"
)

type ContainerManager struct {
	file     string
	lockFile string
}

func NewContainerManager() *ContainerManager {
	return &ContainerManager{
		file:     filepath.Join(os.TempDir(), "goravel_docker.txt"),
		lockFile: filepath.Join(os.TempDir(), "goravel_docker.lock"),
	}
}

func (r *ContainerManager) Add(containerType ContainerType, databaseDriver testing.DatabaseDriver) {
	containerTypeToDatabaseDrivers := r.All()
	containerTypeToDatabaseDrivers[containerType] = append(containerTypeToDatabaseDrivers[containerType], databaseDriver)

	containerTypeToDatabaseConfigs := make(map[ContainerType][]testing.DatabaseConfig)
	for k, v := range containerTypeToDatabaseDrivers {
		containerTypeToDatabaseConfigs[k] = make([]testing.DatabaseConfig, len(v))
		for i, driver := range v {
			containerTypeToDatabaseConfigs[k][i] = driver.Config()
		}
	}

	color.Red().Println("add", r.file, databaseDriver.Config(), containerTypeToDatabaseConfigs)

	f, err := os.OpenFile(r.file, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	content, err := json.NewJson().Marshal(containerTypeToDatabaseConfigs)
	if err != nil {
		panic(err)
	}

	_, err = f.WriteString(string(content))
	if err != nil {
		panic(err)
	}
}

func (r *ContainerManager) All() map[ContainerType][]testing.DatabaseDriver {
	containerTypeToDatabaseDrivers := make(map[ContainerType][]testing.DatabaseDriver)
	if !file.Exists(r.file) {
		return containerTypeToDatabaseDrivers
	}

	f, err := os.OpenFile(r.file, os.O_RDONLY, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}

	var containerTypeToDatabaseConfigs map[ContainerType][]testing.DatabaseConfig
	if err := json.NewJson().Unmarshal(bytes.TrimSpace(content), &containerTypeToDatabaseConfigs); err != nil {
		panic(err)
	}

	if len(containerTypeToDatabaseConfigs) == 0 {
		return containerTypeToDatabaseDrivers
	}

	containerTypeToDatabaseConfigs1 := make(map[ContainerType][]testing.DatabaseConfig)
	for containerType, databaseConfigs := range containerTypeToDatabaseConfigs {
		for _, databaseConfig := range databaseConfigs {
			// If the port is not occupied, provide the container is released.
			if databaseConfig.Port != 0 && !isPortUsing(databaseConfig.Port) {
				continue
			}
			containerTypeToDatabaseConfigs1[containerType] = append(containerTypeToDatabaseConfigs1[containerType], databaseConfig)
			databaseDriver := NewDatabaseDriverByExist(containerType, databaseConfig.ContainerID, databaseConfig.Database, databaseConfig.Username, databaseConfig.Password, databaseConfig.Port)
			containerTypeToDatabaseDrivers[containerType] = append(containerTypeToDatabaseDrivers[containerType], databaseDriver)
		}
	}

	color.Red().Println("all", r.file, containerTypeToDatabaseConfigs1)

	return containerTypeToDatabaseDrivers
}

func (r *ContainerManager) Lock() {
	for {
		if !file.Exists(r.lockFile) {
			break
		}
	}
	if err := file.Create(r.lockFile, ""); err != nil {
		panic(err)
	}
}

func (r *ContainerManager) Unlock() {
	if err := file.Remove(r.lockFile); err != nil {
		panic(err)
	}
}

func (r *ContainerManager) Remove() error {
	if err := file.Remove(r.lockFile); err != nil {
		return err
	}

	return file.Remove(r.file)
}
