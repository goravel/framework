package docker

import (
	"fmt"
	"github.com/goravel/framework/support/str"

	"github.com/goravel/framework/contracts/testing"
	"github.com/goravel/framework/errors"
)

// Define different test model, to improve the local testing speed.
// The minimum model only initials one Sqlite and two Postgres,
// and the normal model initials one Mysql, two Postgres, one Sqlite and one Sqlserver.
const (
	TestModelMinimum = iota
	TestModelNormal

	// Switch this value to control the test model.
	TestModel = TestModelMinimum
)

type ContainerType string

const (
	testDatabase = "goravel"
	testUsername = "goravel"
	testPassword = "Framework!123"

	ContainerTypeMysql     ContainerType = "mysql"
	ContainerTypePostgres  ContainerType = "postgres"
	ContainerTypeSqlite    ContainerType = "sqlite"
	ContainerTypeSqlserver ContainerType = "sqlserver"
	ContainerTypeRedis     ContainerType = "redis"
)

func Mysql() testing.DatabaseDriver {
	return Mysqls(1)[0]
}

func Mysqls(num int) []testing.DatabaseDriver {
	return Database(ContainerTypeMysql, testUsername, testPassword, num)
}

func Postgres() testing.DatabaseDriver {
	return Postgreses(1)[0]
}

func Postgreses(num int) []testing.DatabaseDriver {
	return Database(ContainerTypePostgres, testUsername, testPassword, num)
}

func Sqlserver() testing.DatabaseDriver {
	return Sqlservers(1)[0]
}

func Sqlservers(num int) []testing.DatabaseDriver {
	return Database(ContainerTypeSqlserver, testUsername, testPassword, num)
}

func Sqlite() testing.DatabaseDriver {
	return Sqlites(1)[0]
}

func Sqlites(num int) []testing.DatabaseDriver {
	return Database(ContainerTypeSqlite, testUsername, testPassword, num)
}

func Database(containerType ContainerType, username, password string, num int) []testing.DatabaseDriver {
	if num <= 0 {
		panic(errors.DockerDatabaseContainerCountZero)
	}

	// Get containers from temp file.
	containerManager := NewContainerManager()
	containerManager.Lock()
	defer containerManager.Unlock()

	containerTypeToDatabaseDrivers := containerManager.All()

	var databaseDrivers []testing.DatabaseDriver
	if len(containerTypeToDatabaseDrivers[containerType]) >= num {
		databaseDrivers = containerTypeToDatabaseDrivers[containerType][:num]
	} else {
		databaseDrivers = containerTypeToDatabaseDrivers[containerType]
	}

	// Create new database in the exist docker container
	for i, databaseDriver := range databaseDrivers {
		databaseName := fmt.Sprintf("goravel_%s", str.Random(6))
		if newDatabaseDriver, err := databaseDriver.Database(databaseName); err != nil {
			panic(err)
		} else {
			databaseDrivers[i] = newDatabaseDriver
		}
	}

	// Create new docker container
	driverLength := len(databaseDrivers)
	surplus := num - driverLength
	for i := 0; i < surplus; i++ {
		databaseName := fmt.Sprintf("goravel_%s", str.Random(6))
		databaseDriver := NewDatabaseDriver(containerType, databaseName, username, password)

		if err := databaseDriver.Build(); err != nil {
			panic(err)
		}

		containerManager.Add(containerType, databaseDriver)
		databaseDrivers = append(databaseDrivers, databaseDriver)
	}

	if len(databaseDrivers) != num {
		panic(errors.DockerInsufficientDatabaseContainers.Args(num, len(databaseDrivers)))
	}

	return databaseDrivers
}

func NewDatabaseDriver(containerType ContainerType, database, username, password string) testing.DatabaseDriver {
	switch containerType {
	case ContainerTypeMysql:
		return NewMysqlImpl(database, username, password)
	case ContainerTypePostgres:
		return NewPostgresImpl(database, username, password)
	case ContainerTypeSqlserver:
		return NewSqlserverImpl(database, username, password)
	case ContainerTypeSqlite:
		return NewSqliteImpl(database)
	default:
		panic(errors.DockerUnknownContainerType)
	}
}

func NewDatabaseDriverByExist(containerType ContainerType, containerID, database, username, password string, port int) testing.DatabaseDriver {
	switch containerType {
	case ContainerTypeMysql:
		driver := NewMysqlImpl(database, username, password)
		driver.containerID = containerID
		driver.port = port

		return driver
	case ContainerTypePostgres:
		driver := NewPostgresImpl(database, username, password)
		driver.containerID = containerID
		driver.port = port

		return driver
	case ContainerTypeSqlserver:
		driver := NewSqlserverImpl(database, username, password)
		driver.containerID = containerID
		driver.port = port

		return driver
	case ContainerTypeSqlite:
		return NewSqliteImpl(database)

	default:
		panic(errors.DockerUnknownContainerType)
	}
}

func Stop() error {
	containerManager := NewContainerManager()
	containerTypeToDatabaseDrivers := containerManager.All()

	for _, drivers := range containerTypeToDatabaseDrivers {
		for _, driver := range drivers {
			if err := driver.Stop(); err != nil {
				return err
			}
		}
	}

	return nil
}
