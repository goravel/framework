package schema

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/database/schema/grammars"
	"github.com/goravel/framework/support/env"
)

type SqlserverSchemaSuite struct {
	suite.Suite
	sqlserverSchema *SqlserverSchema
}

func TestSqlserverSchemaSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skip test that using Docker")
	}

	suite.Run(t, &SqlserverSchemaSuite{})
}

func (s *SqlserverSchemaSuite) SetupTest() {
	s.sqlserverSchema = NewSqlserverSchema(grammars.NewSqlserver("goravel_"), nil, "goravel_")
}

func (s *SqlserverSchemaSuite) TestParseSchemaAndTable() {
	tests := []struct {
		reference      string
		expectedSchema string
		expectedTable  string
	}{
		{"public.users", "public", "users"},
		{"users", "", "users"},
	}

	for _, test := range tests {
		schema, table := s.sqlserverSchema.parseSchemaAndTable(test.reference)
		s.Equal(test.expectedSchema, schema)
		s.Equal(test.expectedTable, table)
	}
}
