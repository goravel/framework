package schema

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/database/schema/grammars"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
	"github.com/goravel/framework/support/docker"
	"github.com/goravel/framework/support/env"
)

type PostgresSchemaSuite struct {
	suite.Suite
	mockOrm        *mocksorm.Orm
	postgresSchema *PostgresSchema
	testQuery      *gorm.TestQuery
}

func TestPostgresSchemaSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skip test that using Docker")
	}

	suite.Run(t, &PostgresSchemaSuite{})
}

func (s *PostgresSchemaSuite) SetupTest() {
	postgresDocker := docker.Postgres()
	s.Require().NoError(postgresDocker.Ready())

	s.testQuery = gorm.NewTestQuery(postgresDocker, true)
	s.mockOrm = mocksorm.NewOrm(s.T())
	s.postgresSchema = NewPostgresSchema(grammars.NewPostgres("goravel_"), s.mockOrm, "goravel", "goravel_")
}

// TODO Implement this after implementing create type
func (s *PostgresSchemaSuite) TestGetTypes() {

}

func (s *PostgresSchemaSuite) TestParseSchemaAndTable() {
	tests := []struct {
		reference      string
		expectedSchema string
		expectedTable  string
	}{
		{"public.users", "public", "users"},
		{"users", "goravel", "users"},
	}

	for _, test := range tests {
		schema, table := s.postgresSchema.parseSchemaAndTable(test.reference)
		s.Equal(test.expectedSchema, schema)
		s.Equal(test.expectedTable, table)
	}
}
