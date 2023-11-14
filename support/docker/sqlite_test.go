package docker

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database/orm"
	contractstesting "github.com/goravel/framework/contracts/testing"
	configmocks "github.com/goravel/framework/mocks/config"
	"github.com/goravel/framework/support/env"
)

type SqliteTestSuite struct {
	suite.Suite
	mockConfig *configmocks.Config
	sqlite     *Sqlite
}

func TestSqliteTestSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	suite.Run(t, new(SqliteTestSuite))
}

func (s *SqliteTestSuite) SetupTest() {
	s.mockConfig = &configmocks.Config{}
	s.sqlite = NewSqlite("goravel")
}

func (s *SqliteTestSuite) TestBuild() {
	s.Nil(s.sqlite.Build())
	instance, err := s.sqlite.connect()
	s.Nil(err)
	s.NotNil(instance)

	s.Equal("goravel", s.sqlite.Config().Database)

	res := instance.Exec(`
CREATE TABLE users (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  name varchar(255) NOT NULL
);
`)
	s.Nil(res.Error)

	res = instance.Exec(`
INSERT INTO users (name) VALUES ('goravel');
`)
	s.Nil(res.Error)
	s.Equal(int64(1), res.RowsAffected)

	var count int64
	res = instance.Raw("SELECT count(*) FROM sqlite_master WHERE type='table' and name = 'users';").Scan(&count)
	s.Nil(res.Error)
	s.Equal(int64(1), count)

	s.Nil(s.sqlite.Fresh())

	instance, err = s.sqlite.connect()
	s.Nil(err)
	s.NotNil(instance)

	res = instance.Raw("SELECT count(*) FROM sqlite_master WHERE type='table' and name = 'users';").Scan(&count)
	s.Nil(res.Error)
	s.Equal(int64(0), count)

	s.Nil(s.sqlite.Stop())
}

func (s *SqliteTestSuite) TestImage() {
	image := contractstesting.Image{
		Repository: "sqlite",
	}
	s.sqlite.Image(image)
	s.Equal(&image, s.sqlite.image)
}

func (s *SqliteTestSuite) TestName() {
	s.Equal(orm.DriverSqlite, s.sqlite.Name())
}
