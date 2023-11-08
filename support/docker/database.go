package docker

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	gormio "gorm.io/gorm"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
)

const (
	DbPassword     = "Goravel(!)"
	DbUser         = "goravel"
	DbDatabase     = "goravel"
	MysqlPort      = 9910
	PostgresqlPort = 9920
	SqlserverPort  = 9930
	mysqlPort      = 9700
	postgresqlPort = 9800
	sqlserverPort  = 9900
)

var (
	usingDatabaseNum = 0
	lock             sync.Mutex
)

type Database struct {
	file           *os.File
	User           string
	Password       string
	Database       string
	MysqlPort      int
	PostgresqlPort int
	SqlserverPort  int
}

func InitDatabase() (*Database, error) {
	database := &Database{
		User:           "goravel",
		Password:       "Goravel(!)",
		Database:       "goravel",
		MysqlPort:      mysqlPort + usingDatabaseNum,
		PostgresqlPort: postgresqlPort + usingDatabaseNum,
		SqlserverPort:  sqlserverPort + usingDatabaseNum,
	}

	file, err := os.CreateTemp("", "goravel-docker-composer-*.yml")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if err := os.WriteFile(file.Name(), []byte(Compose{}.Database(database.MysqlPort, database.PostgresqlPort, database.SqlserverPort)), 0755); err != nil {
		return nil, err
	}

	if err := shell(fmt.Sprintf("docker-compose -f %s up --detach --quiet-pull", file.Name())); err != nil {
		return nil, err
	}

	database.file = file
	lock.Lock()
	usingDatabaseNum++
	lock.Unlock()

	return database, nil
}

func (r *Database) Fresh() error {
	if err := r.freshMysql(); err != nil {
		return err
	}
	if err := r.freshPostgresql(); err != nil {
		return err
	}
	if err := r.freshSqlserver(); err != nil {
		return err
	}

	return nil
}

func (r *Database) Stop() error {
	defer func() {
		os.Remove(r.file.Name())
	}()

	return shell(fmt.Sprintf("docker-compose -f %s down", r.file.Name()))
}

func (r *Database) freshMysql() error {
	instance, err := r.connect(contractsorm.DriverMysql)
	if err != nil {
		return fmt.Errorf("connect Mysql error when clearing: %v", err)
	}

	res := instance.Raw("select concat('drop table ',table_name,';') from information_schema.TABLES where table_schema=?;", DbDatabase)
	if res.Error != nil {
		return fmt.Errorf("get tables of Mysql error: %v", res.Error)
	}

	var tables []string
	res = res.Scan(&tables)
	if res.Error != nil {
		return fmt.Errorf("get tables of Mysql error: %v", res.Error)
	}

	for _, table := range tables {
		res = instance.Exec(table)
		if res.Error != nil {
			return fmt.Errorf("drop table %s of Mysql error: %v", table, res.Error)
		}
	}

	return nil
}

func (r *Database) freshPostgresql() error {
	instance, err := r.connect(contractsorm.DriverPostgresql)
	if err != nil {
		return fmt.Errorf("connect Postgres error when clearing: %v", err)
	}

	if res := instance.Exec("DROP SCHEMA public CASCADE;"); res.Error != nil {
		return fmt.Errorf("drop schema of Postgres error: %v", res.Error)
	}

	if res := instance.Exec("CREATE SCHEMA public;"); res.Error != nil {
		return fmt.Errorf("create schema of Postgres error: %v", res.Error)
	}

	return nil
}

func (r *Database) freshSqlserver() error {
	instance, err := r.connect(contractsorm.DriverSqlserver)
	if err != nil {
		return fmt.Errorf("connect Sqlserver error when clearing: %v", err)
	}

	res := instance.Raw("SELECT NAME FROM SYSOBJECTS WHERE TYPE='U';")
	if res.Error != nil {
		return fmt.Errorf("get tables of Sqlserver error: %v", res.Error)
	}

	var tables []string
	res = res.Scan(&tables)
	if res.Error != nil {
		return fmt.Errorf("get tables of Sqlserver error: %v", res.Error)
	}

	for _, table := range tables {
		res = instance.Exec(fmt.Sprintf("drop table %s;", table))
		if res.Error != nil {
			return fmt.Errorf("drop table %s of Sqlserver error: %v", table, res.Error)
		}
	}

	return nil
}

func (r *Database) connect(driver contractsorm.Driver) (*gormio.DB, error) {
	var (
		instance *gormio.DB
		err      error
	)

	// docker compose need time to start
	for i := 0; i < 60; i++ {
		switch driver {
		case contractsorm.DriverMysql:
			instance, err = gormio.Open(mysql.New(mysql.Config{
				DSN: fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", r.User, r.Password, "127.0.0.1", r.MysqlPort, r.Database),
			}))
		case contractsorm.DriverPostgresql:
			instance, err = gormio.Open(postgres.New(postgres.Config{
				DSN: fmt.Sprintf("postgres://%s:%s@%s:%d/%s", r.User, r.Password, "127.0.0.1", r.PostgresqlPort, r.Database),
			}))
		case contractsorm.DriverSqlserver:
			instance, err = gormio.Open(sqlserver.New(sqlserver.Config{
				DSN: fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
					r.User, r.Password, "127.0.0.1", r.SqlserverPort, r.Database),
			}))
		default:
			return nil, fmt.Errorf("not support driver %s", driver)
		}

		if err == nil {
			break
		}

		time.Sleep(2 * time.Second)
	}

	return instance, err
}

func shell(command string) error {
	cmd := exec.Command("/bin/sh", "-c", command)

	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf(fmt.Sprint(err) + ": " + stderr.String())
	}

	return nil
}
