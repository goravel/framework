package docker

import (
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/suite"

	configmocks "github.com/goravel/framework/contracts/config/mocks"
	"github.com/goravel/framework/contracts/database/orm"
)

type PostgresqlTestSuite struct {
	suite.Suite
	mockConfig *configmocks.Config
	postgresql *Postgresql
}

func TestPostgresqlTestSuite(t *testing.T) {
	suite.Run(t, new(PostgresqlTestSuite))
}

func (s *PostgresqlTestSuite) SetupTest() {
	s.mockConfig = configmocks.NewConfig(s.T())
	s.postgresql = &Postgresql{
		config:     s.mockConfig,
		connection: "postgresql",
	}
}

func (s *PostgresqlTestSuite) TestName() {
	s.Equal(orm.DriverPostgresql, s.postgresql.Name())
}

func (s *PostgresqlTestSuite) TestImage() {
	s.mockConfig.On("GetString", "database.connections.postgresql.database").Return("goravel").Once()
	s.mockConfig.On("GetString", "database.connections.postgresql.username").Return("root").Once()
	s.mockConfig.On("GetString", "database.connections.postgresql.password").Return("123123").Once()

	s.Equal(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "latest",
		Env: []string{
			"POSTGRES_USER=root",
			"POSTGRES_PASSWORD=123123",
			"POSTGRES_DB=goravel",
			"listen_addresses = '*'",
		},
	}, s.postgresql.Image())
}
