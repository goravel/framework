package docker

import (
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/suite"

	configmocks "github.com/goravel/framework/contracts/config/mocks"
	"github.com/goravel/framework/contracts/database/orm"
)

type SqliteTestSuite struct {
	suite.Suite
	mockConfig *configmocks.Config
	sqlite     *Sqlite
}

func TestSqliteTestSuite(t *testing.T) {
	suite.Run(t, new(SqliteTestSuite))
}

func (s *SqliteTestSuite) SetupTest() {
	s.mockConfig = configmocks.NewConfig(s.T())
	s.sqlite = &Sqlite{
		config:     s.mockConfig,
		connection: "sqlite",
	}
}

func (s *SqliteTestSuite) TestName() {
	s.Equal(orm.DriverSqlite, s.sqlite.Name())
}

func (s *SqliteTestSuite) TestImage() {
	s.Equal(&dockertest.RunOptions{
		Repository: "nouchka/sqlite3",
		Tag:        "latest",
		Env:        []string{},
	}, s.sqlite.Image())
}
