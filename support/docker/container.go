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
	"github.com/goravel/framework/support/process"
	"github.com/goravel/framework/support/str"
)

type Container struct {
	databaseDriver testing.DatabaseDriver
	file           string
	lockFile       string
	name           string
	username       string
	password       string
}

func NewContainer(databaseDriver testing.DatabaseDriver) *Container {
	return &Container{
		databaseDriver: databaseDriver,
		file:           filepath.Join(os.TempDir(), "goravel_docker.txt"),
		lockFile:       filepath.Join(os.TempDir(), "goravel_docker.lock"),
		name:           databaseDriver.Driver(),
		username:       "goravel",
		password:       "Framework!123",
	}
}

func (r *Container) Build() (testing.DatabaseDriver, error) {
	var (
		isReused bool
		err      error
	)

	r.lock()
	defer r.unlock()

	databaseConfigs, err := r.all()
	if err != nil {
		return nil, err
	}

	// If the port is not occupied, provide the container is released.
	if databaseConfigs != nil {
		if _, exist := databaseConfigs[r.name]; exist && process.IsPortUsing(databaseConfigs[r.name].Port) {
			if err := r.databaseDriver.Reuse(databaseConfigs[r.name].ContainerID, databaseConfigs[r.name].Port); err == nil {
				isReused = true
			}
		}
	}

	if !isReused {
		if err := r.databaseDriver.Build(); err != nil {
			return nil, err
		}

		if err := r.add(); err != nil {
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

func (r *Container) add() error {
	databaseConfigs, err := r.all()
	if err != nil {
		return err
	}

	if databaseConfigs == nil {
		databaseConfigs = make(map[string]testing.DatabaseConfig)
	}
	databaseConfigs[r.name] = r.databaseDriver.Config()
	f, err := os.OpenFile(r.file, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	content, err := json.NewJson().Marshal(databaseConfigs)
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
	databaseConfigs := make(map[string]testing.DatabaseConfig)
	if !file.Exists(r.file) {
		return databaseConfigs, nil
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
	if err := json.NewJson().Unmarshal(content, &databaseConfigs); err != nil {
		return nil, err
	}

	return databaseConfigs, nil
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
