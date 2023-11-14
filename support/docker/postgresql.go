package docker

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	gormio "gorm.io/gorm"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/testing"
)

type Postgresql struct {
	containerID string
	database    string
	host        string
	image       *testing.Image
	password    string
	username    string
	port        int
}

func NewPostgresql(database, username, password string) *Postgresql {
	return &Postgresql{
		database: database,
		host:     "127.0.0.1",
		username: username,
		password: password,
		image: &testing.Image{
			Repository: "postgres",
			Tag:        "latest",
			Env: []string{
				"POSTGRES_USER=" + username,
				"POSTGRES_PASSWORD=" + password,
				"POSTGRES_DB=" + database,
			},
			ExposedPorts: []string{"5432"},
		},
	}
}

func (receiver *Postgresql) Build() error {
	command, exposedPorts := imageToCommand(receiver.image)
	containerID, err := run(command)
	if err != nil {
		return fmt.Errorf("init Postgresql error: %v", err)
	}
	if containerID == "" {
		return fmt.Errorf("no container id return when creating Postgresql docker")
	}

	receiver.containerID = containerID
	receiver.port = getExposedPort(exposedPorts, 5432)

	if _, err := receiver.connect(); err != nil {
		return fmt.Errorf("connect Postgresql error: %v", err)
	}

	return nil
}

func (receiver *Postgresql) Config() testing.DatabaseConfig {
	return testing.DatabaseConfig{
		Host:     receiver.host,
		Port:     receiver.port,
		Database: receiver.database,
		Username: receiver.username,
		Password: receiver.password,
	}
}

func (receiver *Postgresql) Fresh() error {
	instance, err := receiver.connect()
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

func (receiver *Postgresql) Image(image testing.Image) {
	receiver.image = &image
}

func (receiver *Postgresql) Name() orm.Driver {
	return orm.DriverPostgresql
}

func (receiver *Postgresql) Stop() error {
	if _, err := run(fmt.Sprintf("docker stop %s", receiver.containerID)); err != nil {
		return fmt.Errorf("stop Postgresql error: %v", err)
	}

	return nil
}

func (receiver *Postgresql) connect() (*gormio.DB, error) {
	var (
		instance *gormio.DB
		err      error
	)

	// docker compose need time to start
	for i := 0; i < 60; i++ {
		instance, err = gormio.Open(postgres.New(postgres.Config{
			DSN: fmt.Sprintf("postgres://%s:%s@%s:%d/%s", receiver.username, receiver.password, receiver.host, receiver.port, receiver.database),
		}))

		if err == nil {
			break
		}

		time.Sleep(2 * time.Second)
	}

	return instance, err
}
