package docker

import (
	"github.com/goravel/framework/contracts/testing"
)

type ContainerType string

const (
	ContainerTypeMysql     ContainerType = "mysql"
	ContainerTypePostgres  ContainerType = "postgres"
	ContainerTypeSqlite    ContainerType = "sqlite"
	ContainerTypeSqlserver ContainerType = "sqlserver"
	ContainerTypeRedis     ContainerType = "redis"
)

var containers = make(map[ContainerType][]testing.DatabaseDriver)

func Mysql1() testing.DatabaseDriver {
	return Database1(ContainerTypeMysql, 1)[0]
}

func Postgres1() testing.DatabaseDriver {
	return Database1(ContainerTypePostgres, 1)[0]
}

func Sqlserver1() testing.DatabaseDriver {
	return Database1(ContainerTypeSqlserver, 1)[0]
}

func Sqlite1() testing.DatabaseDriver {
	return Database1(ContainerTypeSqlite, 1)[0]
}

func Database1(containerType ContainerType, num int) []testing.DatabaseDriver {
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
			db = NewMysql(database, username, password)
		case ContainerTypePostgres:
			db = NewPostgres(database, username, password)
		case ContainerTypeSqlserver:
			db = NewSqlserver(database, username, password)
		case ContainerTypeSqlite:
			db = NewSqlite(database)
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
