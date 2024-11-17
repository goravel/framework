package docker

import (
	"fmt"
	"time"

	"gorm.io/driver/sqlserver"
	gormio "gorm.io/gorm"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/testing"
)

type SqlserverImpl struct {
	containerID string
	database    string
	host        string
	image       *testing.Image
	password    string
	username    string
	port        int
}

func NewSqlserverImpl(database, username, password string) *SqlserverImpl {
	return &SqlserverImpl{
		database: database,
		host:     "127.0.0.1",
		username: username,
		password: password,
		image: &testing.Image{
			Repository: "mcr.microsoft.com/mssql/server",
			Tag:        "latest",
			Env: []string{
				"ACCEPT_EULA=Y",
				"MSSQL_SA_PASSWORD=" + password,
			},
			ExposedPorts: []string{"1433"},
		},
	}
}

func (r *SqlserverImpl) Build() error {
	command, exposedPorts := imageToCommand(r.image)
	containerID, err := run(command)
	if err != nil {
		return fmt.Errorf("init Sqlserver docker error: %v", err)
	}
	if containerID == "" {
		return fmt.Errorf("no container id return when creating Sqlserver docker")
	}

	r.containerID = containerID
	r.port = getExposedPort(exposedPorts, 1433)

	return nil
}

func (r *SqlserverImpl) Config() testing.DatabaseConfig {
	return testing.DatabaseConfig{
		ContainerID: r.containerID,
		Host:        r.host,
		Port:        r.port,
		Database:    r.database,
		Username:    r.username,
		Password:    r.password,
	}
}

func (r *SqlserverImpl) Database(name string) (testing.DatabaseDriver, error) {
	sqlserverImpl := NewSqlserverImpl(name, r.username, r.password)
	sqlserverImpl.containerID = r.containerID
	sqlserverImpl.port = r.port

	return sqlserverImpl, nil
}

func (r *SqlserverImpl) Driver() database.Driver {
	return database.DriverSqlserver
}

func (r *SqlserverImpl) Fresh() error {
	instance, err := r.connect()
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

func (r *SqlserverImpl) Image(image testing.Image) {
	r.image = &image
}

func (r *SqlserverImpl) Ready() error {
	_, err := r.connect()

	return err
}

func (r *SqlserverImpl) Stop() error {
	if _, err := run(fmt.Sprintf("docker stop %s", r.containerID)); err != nil {
		return fmt.Errorf("stop Sqlserver error: %v", err)
	}

	return nil
}

func (r *SqlserverImpl) connect() (*gormio.DB, error) {
	var (
		instance *gormio.DB
		err      error
	)

	// docker compose need time to start
	for i := 0; i < 100; i++ {
		instance, err = gormio.Open(sqlserver.New(sqlserver.Config{
			DSN: fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=master",
				"sa", r.password, r.host, r.port),
		}))

		if err == nil {
			// Check if database exists
			var exists bool
			query := fmt.Sprintf("SELECT CASE WHEN EXISTS (SELECT * FROM sys.databases WHERE name = '%s') THEN CAST(1 AS BIT) ELSE CAST(0 AS BIT) END", r.database)
			if err := instance.Raw(query).Scan(&exists).Error; err != nil {
				return nil, err
			}

			if !exists {
				// Create User database
				if err := instance.Exec(fmt.Sprintf(`CREATE DATABASE "%s";`, r.database)).Error; err != nil {
					return nil, err
				}

				query = fmt.Sprintf("SELECT 1 FROM sys.server_principals WHERE name = '%s' AND type = 'S'", r.username)
				if err := instance.Raw(query).Scan(&exists).Error; err != nil {
					return nil, err
				}

				if !exists {
					// Create User account
					if err := instance.Exec(fmt.Sprintf("CREATE LOGIN %s WITH PASSWORD = '%s'", r.username, r.password)).Error; err != nil {
						return nil, err
					}
				}

				// Create DB account for User
				if err := instance.Exec(fmt.Sprintf("USE %s; CREATE USER %s FOR LOGIN %s", r.database, r.username, r.username)).Error; err != nil {
					return nil, err
				}

				// Add permission
				if err := instance.Exec(fmt.Sprintf("USE %s; ALTER ROLE db_owner ADD MEMBER %s", r.database, r.username)).Error; err != nil {
					return nil, err
				}
			}

			instance, err = gormio.Open(sqlserver.New(sqlserver.Config{
				DSN: fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
					r.username, r.password, r.host, r.port, r.database),
			}))

			break
		}

		time.Sleep(2 * time.Second)
	}

	return instance, err
}
