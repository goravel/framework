package docker

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	gormio "gorm.io/gorm"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/testing"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/color"
)

type MysqlImpl struct {
	containerID string
	database    string
	host        string
	image       *testing.Image
	password    string
	username    string
	port        int
}

func NewMysqlImpl(database, username, password string) *MysqlImpl {
	env := []string{
		"MYSQL_ROOT_PASSWORD=" + password,
		"MYSQL_DATABASE=" + database,
	}
	if username != "root" {
		env = append(env, "MYSQL_USER="+username)
		env = append(env, "MYSQL_PASSWORD="+password)
	}

	return &MysqlImpl{
		database: database,
		host:     "127.0.0.1",
		username: username,
		password: password,
		image: &testing.Image{
			Repository:   "mysql",
			Tag:          "latest",
			Env:          env,
			ExposedPorts: []string{"3306"},
		},
	}
}

func (r *MysqlImpl) Build() error {
	command, exposedPorts := imageToCommand(r.image)
	containerID, err := run(command)
	if err != nil {
		return fmt.Errorf("init Mysql docker error: %v", err)
	}
	if containerID == "" {
		return errors.DockerMissingContainerId.Args("Mysql")
	}

	r.containerID = containerID
	r.port = getExposedPort(exposedPorts, 3306)

	return nil
}

func (r *MysqlImpl) Config() testing.DatabaseConfig {
	return testing.DatabaseConfig{
		ContainerID: r.containerID,
		Host:        r.host,
		Port:        r.port,
		Database:    r.database,
		Username:    r.username,
		Password:    r.password,
	}
}

func (r *MysqlImpl) Database(name string) (testing.DatabaseDriver, error) {
	// We want to return the DatabaseDriver instance directly, to avoid blocking test cases in different packages,
	// because the test process will be clocked when creating a new container, each package should call the Ready method
	// to check if the container is ready. Returning the DatabaseDriver instance directly will allow to initialize multiple
	// container simultaneously.
	go func() {
		instance, err := r.connect("root")
		if err != nil {
			color.Errorf("connect Mysql error: %v", err)
			return
		}

		res := instance.Exec(fmt.Sprintf(`CREATE DATABASE %s;`, name))
		if res.Error != nil {
			color.Errorf("create Mysql database error: %v", res.Error)
			return
		}

		res = instance.Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON %s.* TO `%s`@`%%`;", name, r.username))
		if res.Error != nil {
			color.Errorf("grant privileges in Mysql database error: %v", res.Error)
		}
	}()

	mysqlImpl := NewMysqlImpl(name, r.username, r.password)
	mysqlImpl.containerID = r.containerID
	mysqlImpl.port = r.port

	return mysqlImpl, nil
}

func (r *MysqlImpl) Driver() database.Driver {
	return database.DriverMysql
}

func (r *MysqlImpl) Fresh() error {
	instance, err := r.connect()
	if err != nil {
		return fmt.Errorf("connect Mysql error when clearing: %v", err)
	}

	res := instance.Raw("select concat('drop table ',table_name,';') from information_schema.TABLES where table_schema=?;", r.database)
	if res.Error != nil {
		return fmt.Errorf("get tables of Mysql error: %v", res.Error)
	}

	var tables []string
	res = res.Scan(&tables)
	if res.Error != nil {
		return fmt.Errorf("get tables of Mysql error: %v", res.Error)
	}

	if res := instance.Exec("SET FOREIGN_KEY_CHECKS=0;"); res.Error != nil {
		return fmt.Errorf("disable foreign key check of Mysql error: %v", res.Error)
	}

	for _, table := range tables {
		res = instance.Exec(table)
		if res.Error != nil {
			return fmt.Errorf("drop table %s of Mysql error: %v", table, res.Error)
		}
	}

	if res := instance.Exec("SET FOREIGN_KEY_CHECKS=1;"); res.Error != nil {
		return fmt.Errorf("enable foreign key check of Mysql error: %v", res.Error)
	}

	return nil
}

func (r *MysqlImpl) Image(image testing.Image) {
	r.image = &image
}

func (r *MysqlImpl) Ready() error {
	_, err := r.connect()

	return err
}

func (r *MysqlImpl) Stop() error {
	if _, err := run(fmt.Sprintf("docker stop %s", r.containerID)); err != nil {
		return fmt.Errorf("stop Mysql error: %v", err)
	}

	return nil
}

func (r *MysqlImpl) connect(username ...string) (*gormio.DB, error) {
	var (
		instance *gormio.DB
		err      error
	)

	useUsername := r.username
	if len(username) > 0 {
		useUsername = username[0]
	}

	// docker compose need time to start
	for i := 0; i < 60; i++ {
		instance, err = gormio.Open(mysql.New(mysql.Config{
			DSN: fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", useUsername, r.password, r.host, r.port, r.database),
		}))

		if err == nil {
			break
		}

		time.Sleep(2 * time.Second)
	}

	return instance, err
}
