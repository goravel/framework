package docker

import (
	"fmt"
	"time"

	"gorm.io/driver/sqlserver"
	gormio "gorm.io/gorm"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/testing"
)

type Sqlserver struct {
	containerID string
	database    string
	host        string
	image       *testing.Image
	password    string
	username    string
	port        int
}

func NewSqlserver(database, username, password string) *Sqlserver {
	return &Sqlserver{
		database: database,
		host:     "127.0.0.1",
		username: username,
		password: password,
		image: &testing.Image{
			Repository: "mcmoe/mssqldocker",
			Tag:        "latest",
			Env: []string{
				"ACCEPT_EULA=Y",
				"MSSQL_DB=" + database,
				"MSSQL_USER=" + username,
				"MSSQL_PASSWORD=" + password,
				"SA_PASSWORD=" + password,
			},
			ExposedPorts: []string{"1433"},
		},
	}
}

func (receiver *Sqlserver) Build() error {
	command, exposedPorts := imageToCommand(receiver.image)
	containerID, err := run(command)
	if err != nil {
		return fmt.Errorf("init Sqlserver docker error: %v", err)
	}
	if containerID == "" {
		return fmt.Errorf("no container id return when creating Sqlserver docker")
	}

	receiver.containerID = containerID
	receiver.port = getExposedPort(exposedPorts, 1433)

	if _, err := receiver.connect(); err != nil {
		return fmt.Errorf("connect Sqlserver docker error: %v", err)
	}

	return nil
}

func (receiver *Sqlserver) Config() testing.DatabaseConfig {
	return testing.DatabaseConfig{
		Host:     receiver.host,
		Port:     receiver.port,
		Database: receiver.database,
		Username: receiver.username,
		Password: receiver.password,
	}
}

func (receiver *Sqlserver) Fresh() error {
	instance, err := receiver.connect()
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

func (receiver *Sqlserver) Image(image testing.Image) {
	receiver.image = &image
}

func (receiver *Sqlserver) Name() orm.Driver {
	return orm.DriverSqlserver
}

func (receiver *Sqlserver) Stop() error {
	if _, err := run(fmt.Sprintf("docker stop %s", receiver.containerID)); err != nil {
		return fmt.Errorf("stop Sqlserver error: %v", err)
	}

	return nil
}

func (receiver *Sqlserver) connect() (*gormio.DB, error) {
	var (
		instance *gormio.DB
		err      error
	)

	// docker compose need time to start
	for i := 0; i < 60; i++ {
		instance, err = gormio.Open(sqlserver.New(sqlserver.Config{
			DSN: fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
				receiver.username, receiver.password, receiver.host, receiver.port, receiver.database),
		}))

		if err == nil {
			break
		}

		time.Sleep(2 * time.Second)
	}

	return instance, err
}
