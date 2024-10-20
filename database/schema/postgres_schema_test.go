package schema

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/database/schema/grammars"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
	"github.com/goravel/framework/support/docker"
	"github.com/goravel/framework/support/env"
)

type PostgresSchemaSuite struct {
	suite.Suite
	mockConfig     *mocksconfig.Config
	mockOrm        *mocksorm.Orm
	schema         *Schema
	postgresSchema *PostgresSchema
	testQuery      *gorm.TestQuery
}

func TestPostgresSchemaSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	suite.Run(t, &PostgresSchemaSuite{})
}

func (s *PostgresSchemaSuite) SetupTest() {
	postgresDocker := docker.Postgres()
	s.testQuery = gorm.NewTestQuery(postgresDocker, true)
	s.mockConfig = s.testQuery.MockConfig()
	s.mockOrm = mocksorm.NewOrm(s.T())
	s.postgresSchema = NewPostgresSchema(s.testQuery.MockConfig(), grammars.NewPostgres(), s.mockOrm)
}

func (s *PostgresSchemaSuite) TestGetSchema() {
	s.mockOrm.EXPECT().Name().Return("postgres").Once()
	s.mockConfig.EXPECT().GetString("database.connections.postgres.search_path").Return("").Once()

	s.Equal(s.postgresSchema.getSchema(), "public")

	s.mockOrm.EXPECT().Name().Return("postgres").Once()
	s.mockConfig.EXPECT().GetString("database.connections.postgres.search_path").Return("goravel").Once()

	s.Equal(s.postgresSchema.getSchema(), "goravel")
}

// TODO Implement this after implementing create type
func (s *PostgresSchemaSuite) TestGetTypes() {

}
