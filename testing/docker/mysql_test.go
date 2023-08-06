package docker

import (
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/suite"

	configmocks "github.com/goravel/framework/contracts/config/mocks"
	"github.com/goravel/framework/contracts/database/orm"
)

type MysqlTestSuite struct {
	suite.Suite
	mockConfig *configmocks.Config
	mysql      *Mysql
}

func TestMysqlTestSuite(t *testing.T) {
	suite.Run(t, new(MysqlTestSuite))
}

func (s *MysqlTestSuite) SetupTest() {
	s.mockConfig = configmocks.NewConfig(s.T())
	s.mysql = &Mysql{
		config:     s.mockConfig,
		connection: "mysql",
	}
}

func (s *MysqlTestSuite) TestName() {
	s.Equal(orm.DriverMysql, s.mysql.Name())
}

func (s *MysqlTestSuite) TestImage() {
	s.mockConfig.On("GetString", "database.connections.mysql.database").Return("goravel").Once()
	s.mockConfig.On("GetString", "database.connections.mysql.username").Return("root").Once()
	s.mockConfig.On("GetString", "database.connections.mysql.password").Return("123123").Once()

	s.Equal(&dockertest.RunOptions{
		Repository: "mysql",
		Tag:        "latest",
		Env: []string{
			"MYSQL_ROOT_PASSWORD=123123",
			"MYSQL_DATABASE=goravel",
		},
	}, s.mysql.Image())

	s.mockConfig.On("GetString", "database.connections.mysql.database").Return("goravel").Once()
	s.mockConfig.On("GetString", "database.connections.mysql.username").Return("goravel").Once()
	s.mockConfig.On("GetString", "database.connections.mysql.password").Return("123123").Once()

	s.Equal(&dockertest.RunOptions{
		Repository: "mysql",
		Tag:        "latest",
		Env: []string{
			"MYSQL_ROOT_PASSWORD=123123",
			"MYSQL_DATABASE=goravel",
			"MYSQL_USER=goravel",
			"MYSQL_PASSWORD=123123",
		},
	}, s.mysql.Image())
}
