package docker

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	gormio "gorm.io/gorm"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/testing"
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

func (receiver *MysqlImpl) Build() error {
	command, exposedPorts := imageToCommand(receiver.image)
	containerID, err := run(command)
	if err != nil {
		return fmt.Errorf("init Mysql docker error: %v", err)
	}
	if containerID == "" {
		return fmt.Errorf("no container id return when creating Mysql docker")
	}

	receiver.containerID = containerID
	receiver.port = getExposedPort(exposedPorts, 3306)

	if _, err := receiver.connect(); err != nil {
		return fmt.Errorf("connect Mysql docker error: %v", err)
	}

	return nil
}

func (receiver *MysqlImpl) Config() testing.DatabaseConfig {
	return testing.DatabaseConfig{
		Host:     receiver.host,
		Port:     receiver.port,
		Database: receiver.database,
		Username: receiver.username,
		Password: receiver.password,
	}
}

func (receiver *MysqlImpl) Driver() database.Driver {
	return database.DriverMysql
}

func (receiver *MysqlImpl) Fresh() error {
	instance, err := receiver.connect()
	if err != nil {
		return fmt.Errorf("connect Mysql error when clearing: %v", err)
	}

	res := instance.Raw("select concat('drop table ',table_name,';') from information_schema.TABLES where table_schema=?;", testDatabase)
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

func (receiver *MysqlImpl) Image(image testing.Image) {
	receiver.image = &image
}

func (receiver *MysqlImpl) Stop() error {
	if _, err := run(fmt.Sprintf("docker stop %s", receiver.containerID)); err != nil {
		return fmt.Errorf("stop Mysql error: %v", err)
	}

	return nil
}

func (receiver *MysqlImpl) connect() (*gormio.DB, error) {
	var (
		instance *gormio.DB
		err      error
	)

	// docker compose need time to start
	for i := 0; i < 60; i++ {
		instance, err = gormio.Open(mysql.New(mysql.Config{
			DSN: fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", receiver.username, receiver.password, receiver.host, receiver.port, receiver.database),
		}))

		if err == nil {
			break
		}

		time.Sleep(2 * time.Second)
	}

	return instance, err
}
