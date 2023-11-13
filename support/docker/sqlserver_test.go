package docker

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database/orm"
	contractstesting "github.com/goravel/framework/contracts/testing"
	configmocks "github.com/goravel/framework/mocks/config"
)

type SqlserverTestSuite struct {
	suite.Suite
	mockConfig *configmocks.Config
	sqlserver  *Sqlserver
}

func TestSqlserverTestSuite(t *testing.T) {
	suite.Run(t, new(SqlserverTestSuite))
}

func (s *SqlserverTestSuite) SetupTest() {
	s.mockConfig = &configmocks.Config{}
	s.sqlserver = NewSqlserver("goravel", "goravel", "Goravel123")
}

func (s *SqlserverTestSuite) TestConfig() {
	s.Equal("127.0.0.1", s.sqlserver.Config().Host)
	s.Equal("goravel", s.sqlserver.Config().Database)
	s.Equal("goravel", s.sqlserver.Config().Username)
	s.Equal("Goravel123", s.sqlserver.Config().Password)
}

func (s *SqlserverTestSuite) TestImage() {
	image := contractstesting.Image{
		Repository: "sqlserver",
	}
	s.sqlserver.Image(image)
	s.Equal(&image, s.sqlserver.image)
}

func (s *SqlserverTestSuite) TestName() {
	s.Equal(orm.DriverSqlserver, s.sqlserver.Name())
}

func (s *SqlserverTestSuite) TestFresh() {
	s.Nil(s.sqlserver.Build())
	instance, err := s.sqlserver.connect()
	s.Nil(err)
	s.NotNil(instance)

	res := instance.Exec(`
	CREATE TABLE users (
	 id bigint NOT NULL IDENTITY(1,1),
	 name varchar(255) NOT NULL,
	 PRIMARY KEY (id)
	);
	`)
	s.Nil(res.Error)

	res = instance.Exec(`
	INSERT INTO users (name) VALUES ('goravel');
	`)
	s.Nil(res.Error)
	s.Equal(int64(1), res.RowsAffected)

	var count int64
	res = instance.Raw(`
	SELECT count(*) FROM sys.tables WHERE name = 'users';
	`).Scan(&count)
	s.Nil(res.Error)
	s.Equal(int64(1), count)

	s.Nil(s.sqlserver.Fresh())

	res = instance.Raw(`
	SELECT count(*) FROM sys.tables WHERE name = 'users';
	`).Scan(&count)
	s.Nil(res.Error)
	s.Equal(int64(0), count)

	s.Nil(s.sqlserver.Stop())
}
