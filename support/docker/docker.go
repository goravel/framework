package docker

import (
	"fmt"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/testing"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/str"
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

func Mysql() testing.DatabaseDriver {
	return Mysqls(1)[0]
}

func Mysqls(num int) []testing.DatabaseDriver {
	return Database(ContainerTypeMysql, num)
}

func Postgres() testing.DatabaseDriver {
	return Postgreses(1)[0]
}

func Postgreses(num int) []testing.DatabaseDriver {
	return Database(ContainerTypePostgres, num)
}

func Sqlserver() testing.DatabaseDriver {
	return Sqlservers(1)[0]
}

func Sqlservers(num int) []testing.DatabaseDriver {
	return Database(ContainerTypeSqlserver, num)
}

func Sqlite() testing.DatabaseDriver {
	return Sqlites(1)[0]
}

func Sqlites(num int) []testing.DatabaseDriver {
	return Database(ContainerTypeSqlite, num)
}

func Ready(drivers ...testing.DatabaseDriver) error {
	for _, driver := range drivers {
		if err := driver.Ready(); err != nil {
			return err
		}
	}

	return nil
}

func Database(containerType ContainerType, num int) []testing.DatabaseDriver {
	if num <= 0 {
		panic(errors.DockerDatabaseContainerCountZero)
	}

	containerManager := NewContainerManager()
	databaseDriver, err := containerManager.Get(containerType)
	if err != nil {
		panic(err)
	}

	var databaseDrivers []testing.DatabaseDriver

	// Create new database in the exist docker container
	for i := 0; i < num; i++ {
		// Sqlite should be a new database, so we can return it directly.
		if i == 0 && databaseDriver.Driver() == database.DriverSqlite {
			databaseDrivers = append(databaseDrivers, databaseDriver)
		} else {
			databaseName := fmt.Sprintf("goravel_%s", str.Random(6))
			if newDatabaseDriver, err := databaseDriver.Database(databaseName); err != nil {
				panic(err)
			} else {
				databaseDrivers = append(databaseDrivers, newDatabaseDriver)
			}
		}
	}

	if len(databaseDrivers) != num {
		panic(errors.DockerInsufficientDatabaseContainers)
	}

	return databaseDrivers
}
