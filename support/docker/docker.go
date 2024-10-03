package docker

import (
	"fmt"

	"github.com/goravel/framework/contracts/testing"
)

// Define different test model, to improve the local testing speed.
// The minimum model only initials one Sqlite and two Postgres,
// and the normal model initials one Mysql, two Postgres, one Sqlite and one Sqlserver.
const (
	TestModelMinimum = iota
	TestModelNormal

	// Switch this value to control the test model.
	TestModel = TestModelNormal
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

var containers = make(map[ContainerType][]testing.DatabaseDriver)

func Mysql() testing.DatabaseDriver {
	return Mysqls(1)[0]
}

func Mysqls(num int) []testing.DatabaseDriver {
	return Database(ContainerTypeMysql, testDatabase, testUsername, testPassword, num)
}

func Postgres() testing.DatabaseDriver {
	return Postgreses(1)[0]
}

func Postgreses(num int) []testing.DatabaseDriver {
	return Database(ContainerTypePostgres, testDatabase, testUsername, testPassword, num)
}

func Sqlserver() testing.DatabaseDriver {
	return Sqlservers(1)[0]
}

func Sqlservers(num int) []testing.DatabaseDriver {
	return Database(ContainerTypeSqlserver, testDatabase, testUsername, testPassword, num)
}

func Sqlite() testing.DatabaseDriver {
	return Sqlites(1)[0]
}

func Sqlites(num int) []testing.DatabaseDriver {
	return Database(ContainerTypeSqlite, testDatabase, testUsername, testPassword, num)
}

func Database(containerType ContainerType, database, username, password string, num int) []testing.DatabaseDriver {
	if num <= 0 {
		panic("the number of database container must be greater than 0")
	}

	var drivers []testing.DatabaseDriver
	if len(containers[containerType]) >= num {
		drivers = containers[containerType][:num]
	} else {
		drivers = containers[containerType]
	}

	newDatabase := database
	driverLength := len(drivers)
	surplus := num - driverLength
	for i := 0; i < surplus; i++ {
		if containerType == ContainerTypeSqlite {
			newDatabase = fmt.Sprintf("%s%d", database, driverLength+i)
		}
		databaseDriver := DatabaseDriver(containerType, newDatabase, username, password)

		if err := databaseDriver.Build(); err != nil {
			panic(err)
		}

		containers[containerType] = append(containers[containerType], databaseDriver)
		drivers = append(drivers, databaseDriver)
	}

	if len(drivers) != num {
		panic(fmt.Sprintf("the number of database container is not enough, expect: %d, got: %d", num, len(drivers)))
	}

	for _, driver := range drivers {
		if err := driver.Fresh(); err != nil {
			panic(err)
		}
	}

	return drivers
}

func DatabaseDriver(containerType ContainerType, database, username, password string) testing.DatabaseDriver {
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
		panic("unknown container type")
	}
}

func Stop() error {
	for _, drivers := range containers {
		for _, driver := range drivers {
			if err := driver.Stop(); err != nil {
				return err
			}
		}
	}

	return nil
}
