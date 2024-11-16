package docker

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	gormio "gorm.io/gorm"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/testing"
	"github.com/goravel/framework/support/color"
)

type PostgresImpl struct {
	containerID string
	database    string
	host        string
	image       *testing.Image
	password    string
	username    string
	port        int
}

func NewPostgresImpl(database, username, password string) *PostgresImpl {
	return &PostgresImpl{
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

func (r *PostgresImpl) Build() error {
	command, exposedPorts := imageToCommand(r.image)
	containerID, err := run(command)
	if err != nil {
		return fmt.Errorf("init Postgres error: %v", err)
	}
	if containerID == "" {
		return fmt.Errorf("no container id return when creating Postgres docker")
	}

	r.containerID = containerID
	r.port = getExposedPort(exposedPorts, 5432)

	return nil
}

func (r *PostgresImpl) Config() testing.DatabaseConfig {
	return testing.DatabaseConfig{
		ContainerID: r.containerID,
		Host:        r.host,
		Port:        r.port,
		Database:    r.database,
		Username:    r.username,
		Password:    r.password,
	}
}

func (r *PostgresImpl) Database(name string) (testing.DatabaseDriver, error) {
	go func() {
		gormDB, err := r.connect()
		if err != nil {
			color.Errorf("connect Postgres error: %v", err)
			return
		}

		res := gormDB.Exec(fmt.Sprintf(`CREATE DATABASE "%s";`, name))
		if res.Error != nil {
			color.Errorf("create Postgres database error: %v", res.Error)
		}

		if err := r.close(gormDB); err != nil {
			color.Errorf("close Postgres connection error: %v", err)
		}
	}()

	postgresImpl := NewPostgresImpl(name, r.username, r.password)
	postgresImpl.containerID = r.containerID
	postgresImpl.port = r.port

	return postgresImpl, nil
}

func (r *PostgresImpl) Driver() database.Driver {
	return database.DriverPostgres
}

func (r *PostgresImpl) Fresh() error {
	gormDB, err := r.connect()
	if err != nil {
		return fmt.Errorf("connect Postgres error when clearing: %v", err)
	}

	if res := gormDB.Exec("DROP SCHEMA public CASCADE;"); res.Error != nil {
		return fmt.Errorf("drop schema of Postgres error: %v", res.Error)
	}

	if res := gormDB.Exec("CREATE SCHEMA public;"); res.Error != nil {
		return fmt.Errorf("create schema of Postgres error: %v", res.Error)
	}

	return r.close(gormDB)
}

func (r *PostgresImpl) Image(image testing.Image) {
	r.image = &image
}

func (r *PostgresImpl) Ready() error {
	gormDB, err := r.connect()
	if err != nil {
		return err
	}

	return r.close(gormDB)
}

func (r *PostgresImpl) Stop() error {
	if _, err := run(fmt.Sprintf("docker stop %s", r.containerID)); err != nil {
		return fmt.Errorf("stop Postgres error: %v", err)
	}

	return nil
}

func (r *PostgresImpl) connect() (*gormio.DB, error) {
	var (
		instance *gormio.DB
		err      error
	)

	// docker compose need time to start
	for i := 0; i < 60; i++ {
		instance, err = gormio.Open(postgres.New(postgres.Config{
			DSN: fmt.Sprintf("postgres://%s:%s@%s:%d/%s", r.username, r.password, r.host, r.port, r.database),
		}))

		if err == nil {
			break
		}

		time.Sleep(2 * time.Second)
	}

	return instance, err
}

func (r *PostgresImpl) close(gormDB *gormio.DB) error {
	db, err := gormDB.DB()
	if err != nil {
		return err
	}

	return db.Close()
}
