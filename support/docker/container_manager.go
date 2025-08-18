package docker

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/goravel/framework/contracts/testing"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/foundation/json"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
)

type ContainerType string

const (
	ContainerTypeMysql     ContainerType = "mysql"
	ContainerTypePostgres  ContainerType = "postgres"
	ContainerTypeSqlite    ContainerType = "sqlite"
	ContainerTypeSqlserver ContainerType = "sqlserver"
)

type ContainerManager struct {
	file     string
	lockFile string
	username string
	password string
}

func NewContainerManager() *ContainerManager {
	return &ContainerManager{
		file:     filepath.Join(os.TempDir(), "goravel_docker.txt"),
		lockFile: filepath.Join(os.TempDir(), "goravel_docker.lock"),
		username: "goravel",
		password: "Framework!123",
	}
}

func (r *ContainerManager) Create(containerType ContainerType, database, username, password string) (testing.DatabaseDriver, error) {
	var databaseDriver testing.DatabaseDriver

	switch containerType {
	case ContainerTypeMysql:
		databaseDriver = NewMysqlImpl(database, username, password)
	case ContainerTypePostgres:
		databaseDriver = NewPostgresImpl(database, username, password)
	case ContainerTypeSqlserver:
		databaseDriver = NewSqlserverImpl(database, username, password)
	case ContainerTypeSqlite:
		databaseDriver = NewSqliteImpl(database)
	default:
		return nil, errors.DockerUnknownContainerType
	}

	if err := databaseDriver.Build(); err != nil {
		return nil, err
	}

	return databaseDriver, nil
}

func (r *ContainerManager) Get(containerType ContainerType) (testing.DatabaseDriver, error) {
	var (
		databaseDriver testing.DatabaseDriver
		err            error
	)

	r.lock()
	defer r.unlock()

	if containerType != ContainerTypeSqlite {
		containerTypeToDatabaseConfig, err := r.all()
		if err != nil {
			return nil, err
		}

		// If the port is not occupied, provide the container is released.
		if containerTypeToDatabaseConfig != nil {
			if _, exist := containerTypeToDatabaseConfig[containerType]; exist && isPortUsing(containerTypeToDatabaseConfig[containerType].Port) {
				databaseDriver = r.databaseConfigToDatabaseDriver(containerType, containerTypeToDatabaseConfig[containerType])
			}
		}
	}
	if databaseDriver == nil {
		database := fmt.Sprintf("goravel_%s", str.Random(6))
		databaseDriver, err = r.Create(containerType, database, r.username, r.password)
		if err != nil {
			return nil, err
		}

		// Sqlite doesn't need to create a docker container, so it doesn't need to be added to the file, and create it every time.
		if containerType != ContainerTypeSqlite {
			if err := r.add(containerType, databaseDriver); err != nil {
				return nil, err
			}
		}
	}

	return databaseDriver, nil
}

func (r *ContainerManager) Remove() error {
	if err := file.Remove(r.lockFile); err != nil {
		return err
	}

	return file.Remove(r.file)
}

func (r *ContainerManager) add(containerType ContainerType, databaseDriver testing.DatabaseDriver) error {
	containerTypeToDatabaseConfig, err := r.all()
	if err != nil {
		return err
	}

	if containerTypeToDatabaseConfig == nil {
		containerTypeToDatabaseConfig = make(map[ContainerType]testing.DatabaseConfig)
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

func (r *ContainerManager) all() (map[ContainerType]testing.DatabaseConfig, error) {
	containerTypeToDatabaseConfig := make(map[ContainerType]testing.DatabaseConfig)
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

func (r *ContainerManager) databaseConfigToDatabaseDriver(containerType ContainerType, databaseConfig testing.DatabaseConfig) testing.DatabaseDriver {
	switch containerType {
	case ContainerTypeMysql:
		driver := NewMysqlImpl(databaseConfig.Database, databaseConfig.Username, databaseConfig.Password)
		driver.containerID = databaseConfig.ContainerID
		driver.port = databaseConfig.Port

		return driver
	case ContainerTypePostgres:
		driver := NewPostgresImpl(databaseConfig.Database, databaseConfig.Username, databaseConfig.Password)
		driver.containerID = databaseConfig.ContainerID
		driver.port = databaseConfig.Port

		return driver
	case ContainerTypeSqlserver:
		driver := NewSqlserverImpl(databaseConfig.Database, databaseConfig.Username, databaseConfig.Password)
		driver.containerID = databaseConfig.ContainerID
		driver.port = databaseConfig.Port

		return driver
	case ContainerTypeSqlite:
		return NewSqliteImpl(databaseConfig.Database)
	default:
		panic(errors.DockerUnknownContainerType)
	}
}

func (r *ContainerManager) lock() {
	for file.Exists(r.lockFile) {
		time.Sleep(1 * time.Second)
	}
	if err := file.Create(r.lockFile, ""); err != nil {
		panic(err)
	}
}

func (r *ContainerManager) unlock() {
	if err := file.Remove(r.lockFile); err != nil {
		panic(err)
	}
}
