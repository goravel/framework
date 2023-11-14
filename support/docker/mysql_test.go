package docker

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database/orm"
	contractstesting "github.com/goravel/framework/contracts/testing"
	configmocks "github.com/goravel/framework/mocks/config"
	"github.com/goravel/framework/support/env"
)

type MysqlTestSuite struct {
	suite.Suite
	mockConfig *configmocks.Config
	mysql      *Mysql
}

func TestMysqlTestSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	suite.Run(t, new(MysqlTestSuite))
}

func (s *MysqlTestSuite) SetupTest() {
	s.mockConfig = &configmocks.Config{}
	s.mysql = NewMysql("goravel", "goravel", "goravel")
}

func (s *MysqlTestSuite) TestBuild() {
	s.Nil(s.mysql.Build())
	instance, err := s.mysql.connect()
	s.Nil(err)
	s.NotNil(instance)

	s.Equal("127.0.0.1", s.mysql.Config().Host)
	s.Equal("goravel", s.mysql.Config().Database)
	s.Equal("goravel", s.mysql.Config().Username)
	s.Equal("goravel", s.mysql.Config().Password)
	s.True(s.mysql.Config().Port > 0)

	res := instance.Exec(`
CREATE TABLE users (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  name varchar(255) NOT NULL,
  PRIMARY KEY (id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`)
	s.Nil(res.Error)

	res = instance.Exec(`
INSERT INTO users (name) VALUES ('goravel');
`)
	s.Nil(res.Error)
	s.Equal(int64(1), res.RowsAffected)

	var count int64
	res = instance.Raw(fmt.Sprintf("SELECT count(*) FROM information_schema.tables WHERE table_schema = '%s' and table_name = 'users';", s.mysql.Config().Database)).Scan(&count)
	s.Nil(res.Error)
	s.Equal(int64(1), count)

	s.Nil(s.mysql.Fresh())

	res = instance.Raw(fmt.Sprintf("SELECT count(*) FROM information_schema.tables WHERE table_schema = '%s' and table_name = 'users';", s.mysql.Config().Database)).Scan(&count)
	s.Nil(res.Error)
	s.Equal(int64(0), count)

	s.Nil(s.mysql.Stop())
}

func (s *MysqlTestSuite) TestImage() {
	image := contractstesting.Image{
		Repository: "mysql",
	}
	s.mysql.Image(image)
	s.Equal(&image, s.mysql.image)
}

func (s *MysqlTestSuite) TestName() {
	s.Equal(orm.DriverMysql, s.mysql.Name())
}
