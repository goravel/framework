package docker

import (
	"fmt"

	"github.com/goravel/framework/contracts/testing"
)

type ContainerType string

const (
	password = "Goravel123"
	username = "goravel"
	database = "goravel"

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

func Database(containerType ContainerType, num int) []testing.DatabaseDriver {
	var drivers []testing.DatabaseDriver
	if len(containers[containerType]) >= num {
		drivers = containers[containerType][:num]
	} else {
		drivers = containers[containerType]
	}

	for i := 0; i < num-len(drivers); i++ {
		var db testing.DatabaseDriver

		switch containerType {
		case ContainerTypeMysql:
			db = NewMysqlImpl(database, username, password)
		case ContainerTypePostgres:
			db = NewPostgresImpl(database, username, password)
		case ContainerTypeSqlserver:
			db = NewSqlserverImpl(database, username, password)
		case ContainerTypeSqlite:
			db = NewSqliteImpl(fmt.Sprintf("%s%d", database, i))
		}

		if err := db.Build(); err != nil {
			panic(err)
		}

		containers[containerType] = append(containers[containerType], db)
		drivers = append(drivers, db)
	}

	for _, driver := range drivers {
		if err := driver.Fresh(); err != nil {
			panic(err)
		}
	}

	return drivers
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
