package docker

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/goravel/framework/contracts/testing"
	"github.com/goravel/framework/foundation/json"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
)

type Container struct {
	databaseDriver testing.DatabaseDriver
	file           string
	lockFile       string
	username       string
	password       string
}

func NewContainer(databaseDriver testing.DatabaseDriver) *Container {
	return &Container{
		databaseDriver: databaseDriver,
		file:           filepath.Join(os.TempDir(), "goravel_docker.txt"),
		lockFile:       filepath.Join(os.TempDir(), "goravel_docker.lock"),
		username:       "goravel",
		password:       "Framework!123",
	}
}

func (r *Container) Build() (testing.DatabaseDriver, error) {
	var (
		isReused bool
		err      error

		driverName = r.databaseDriver.Driver()
	)

	r.lock()
	defer r.unlock()

	containerTypeToDatabaseConfig, err := r.all()
	if err != nil {
		return nil, err
	}

	// If the port is not occupied, provide the container is released.
	if containerTypeToDatabaseConfig != nil {
		if _, exist := containerTypeToDatabaseConfig[driverName]; exist && isPortUsing(containerTypeToDatabaseConfig[driverName].Port) {
			if err := r.databaseDriver.Reuse(containerTypeToDatabaseConfig[driverName].ContainerID, containerTypeToDatabaseConfig[driverName].Port); err == nil {
				isReused = true
			}
		}
	}

	if !isReused {
		if err := r.databaseDriver.Build(); err != nil {
			return nil, err
		}

		if err := r.add(driverName, r.databaseDriver); err != nil {
			return nil, err
		}
	}

	database := fmt.Sprintf("goravel_%s", str.Random(6))

	return r.databaseDriver.Database(database)
}

func (r *Container) Builds(num int) ([]testing.DatabaseDriver, error) {
	var databaseDrivers []testing.DatabaseDriver
	for i := 0; i < num; i++ {
		databaseDriver, err := r.Build()
		if err != nil {
			return nil, err
		}

		databaseDrivers = append(databaseDrivers, databaseDriver)
	}

	return databaseDrivers, nil
}

func (r *Container) Ready() error {
	return r.databaseDriver.Ready()
}

func (r *Container) Remove() error {
	if err := file.Remove(r.lockFile); err != nil {
		return err
	}

	return file.Remove(r.file)
}

func (r *Container) add(containerType string, databaseDriver testing.DatabaseDriver) error {
	containerTypeToDatabaseConfig, err := r.all()
	if err != nil {
		return err
	}

	if containerTypeToDatabaseConfig == nil {
		containerTypeToDatabaseConfig = make(map[string]testing.DatabaseConfig)
	}
	containerTypeToDatabaseConfig[containerType] = databaseDriver.Config()
	f, err := os.OpenFile(r.file, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	content, err := json.NewJson().Marshal(containerTypeToDatabaseConfig)
	if err != nil {
		return err
	}

	_, err = f.WriteString(string(content))
	if err != nil {
		return err
	}

	return nil
}

func (r *Container) all() (map[string]testing.DatabaseConfig, error) {
	containerTypeToDatabaseConfig := make(map[string]testing.DatabaseConfig)
	if !file.Exists(r.file) {
		return containerTypeToDatabaseConfig, nil
	}

	f, err := os.OpenFile(r.file, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	if err := json.NewJson().Unmarshal(content, &containerTypeToDatabaseConfig); err != nil {
		return nil, err
	}

	return containerTypeToDatabaseConfig, nil
}

func (r *Container) lock() {
	for {
		if !file.Exists(r.lockFile) {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if err := file.Create(r.lockFile, ""); err != nil {
		panic(err)
	}
}

func (r *Container) unlock() {
	if err := file.Remove(r.lockFile); err != nil {
		panic(err)
	}
}
