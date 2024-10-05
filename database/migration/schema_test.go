package migration

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/database/gorm"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
	"github.com/goravel/framework/support/docker"
	"github.com/goravel/framework/support/env"
)

type SchemaSuite struct {
	suite.Suite
	driverToTestQuery map[database.Driver]*gorm.TestQuery
}

func TestSchemaSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	suite.Run(t, &SchemaSuite{})
}

func (s *SchemaSuite) SetupTest() {
	postgresDocker := docker.Postgres()
	postgresQuery := gorm.NewTestQuery(postgresDocker, true)
	s.driverToTestQuery = map[database.Driver]*gorm.TestQuery{
		database.DriverPostgres: postgresQuery,
	}
}

func (s *SchemaSuite) TestDropIfExists() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver.String(), func() {
			schema, mockOrm := initSchema(s.T(), testQuery)

			table := "drop_if_exists"

			mockOrm.EXPECT().Connection(schema.connection).Return(mockOrm).Twice()
			mockOrm.EXPECT().Query().Return(testQuery.Query()).Twice()
			s.NoError(schema.DropIfExists(table))
			s.NoError(schema.Create(table, func(table migration.Blueprint) {
				table.String("name")
			}))

			mockOrm.EXPECT().Query().Return(testQuery.Query()).Once()
			s.True(schema.HasTable(table))

			mockOrm.EXPECT().Connection(schema.connection).Return(mockOrm).Once()
			mockOrm.EXPECT().Query().Return(testQuery.Query()).Once()
			s.NoError(schema.DropIfExists(table))

			mockOrm.EXPECT().Query().Return(testQuery.Query()).Once()
			s.False(schema.HasTable(table))
		})
	}
}

func (s *SchemaSuite) TestTable() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver.String(), func() {
			schema, mockOrm := initSchema(s.T(), testQuery)

			mockOrm.EXPECT().Connection(schema.connection).Return(mockOrm).Once()
			mockOrm.EXPECT().Query().Return(testQuery.Query()).Times(3)

			err := schema.Create("changes", func(table migration.Blueprint) {
				table.String("name")
			})
			s.NoError(err)
			s.True(schema.HasTable("changes"))

			tables, err := schema.GetTables()
			s.NoError(err)
			s.Greater(len(tables), 0)

			// Open this after implementing other methods
			//s.Require().True(schema.HasColumn("changes", "name"))
			//columns, err := schema.GetColumns("changes")
			//s.Require().Nil(err)
			//for _, column := range columns {
			//	if column.Name == "name" {
			//		s.False(column.AutoIncrement)
			//		s.Empty(column.Collation)
			//		s.Empty(column.Comment)
			//		s.Empty(column.Default)
			//		s.False(column.Nullable)
			//		s.Equal("character varying(255)", column.Type)
			//		s.Equal("varchar", column.TypeName)
			//	}
			//}
			//
			//err = schema.Table("changes", func(table migration.Blueprint) {
			//	table.Integer("age")
			//	table.String("name").Comment("This is a name column").Default("goravel").Change()
			//})
			//s.Nil(err)
			//s.True(schema.HasTable("changes"))
			//s.Require().True(schema.HasColumns("changes", []string{"name", "age"}))
			//columns, err = schema.GetColumns("changes")
			//s.Require().Nil(err)
			//for _, column := range columns {
			//	if column.Name == "name" {
			//		s.False(column.AutoIncrement)
			//		s.Empty(column.Collation)
			//		s.Equal("This is a name column", column.Comment)
			//		s.Equal("'goravel'::character varying", column.Default)
			//		s.False(column.Nullable)
			//		s.Equal("character varying(255)", column.Type)
			//		s.Equal("varchar", column.TypeName)
			//	}
			//	if column.Name == "age" {
			//		s.False(column.AutoIncrement)
			//		s.Empty(column.Collation)
			//		s.Empty(column.Comment)
			//		s.Empty(column.Default)
			//		s.False(column.Nullable)
			//		s.Equal("integer", column.Type)
			//		s.Equal("int4", column.TypeName)
			//	}
			//}
		})
	}
}

func initSchema(t *testing.T, testQuery *gorm.TestQuery) (*Schema, *mocksorm.Orm) {
	mockOrm := mocksorm.NewOrm(t)
	schema := NewSchema(testQuery.MockConfig(), testQuery.Docker().Driver().String(), nil, mockOrm)

	return schema, mockOrm
}
