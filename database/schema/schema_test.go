package schema

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	contractsschema "github.com/goravel/framework/contracts/database/schema"
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

func (s *SchemaSuite) TestCreate_DropIfExists_HasTable() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver.String(), func() {
			schema, mockOrm := GetTestSchema(s.T(), testQuery)
			table := "drop_if_exists"
			mockTransaction(mockOrm, testQuery)

			s.NoError(schema.DropIfExists(table))

			mockTransaction(mockOrm, testQuery)

			s.NoError(schema.Create(table, func(table contractsschema.Blueprint) {
				table.String("name")
			}))

			mockOrm.EXPECT().Query().Return(testQuery.Query()).Once()

			s.True(schema.HasTable(table))

			mockTransaction(mockOrm, testQuery)

			s.NoError(schema.DropIfExists(table))

			mockOrm.EXPECT().Query().Return(testQuery.Query()).Once()
			s.False(schema.HasTable(table))
		})
	}
}

func (s *SchemaSuite) TestDropAllTables() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver.String(), func() {
			schema, mockOrm := GetTestSchema(s.T(), testQuery)
			table := "drop_all_tables"

			mockTransaction(mockOrm, testQuery)

			s.NoError(schema.Create(table, func(table contractsschema.Blueprint) {
				table.String("name")
			}))

			mockOrm.EXPECT().Query().Return(testQuery.Query()).Once()

			s.True(schema.HasTable(table))

			mockOrm.EXPECT().Name().Return("postgres").Once()
			testQuery.MockConfig().EXPECT().GetString("database.connections.postgres.search_path").Return("").Once()
			mockOrm.EXPECT().Query().Return(testQuery.Query()).Twice()

			s.NoError(schema.DropAllTables())

			mockOrm.EXPECT().Query().Return(testQuery.Query()).Once()
			s.False(schema.HasTable(table))
		})
	}
}

// TODO Implement this after implementing create type
func (s *SchemaSuite) TestDropAllTypes() {

}

// TODO Implement this after implementing create view
func (s *SchemaSuite) TestDropAllViews() {

}

func (s *SchemaSuite) TestTable_GetTables() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver.String(), func() {
			schema, mockOrm := GetTestSchema(s.T(), testQuery)
			mockTransaction(mockOrm, testQuery)

			s.NoError(schema.Create("changes", func(table contractsschema.Blueprint) {
				table.String("name")
			}))

			mockOrm.EXPECT().Query().Return(testQuery.Query()).Once()

			s.True(schema.HasTable("changes"))

			mockOrm.EXPECT().Query().Return(testQuery.Query()).Once()

			tables, err := schema.GetTables()
			s.NoError(err)
			s.Len(tables, 1)

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

// TODO Implement this after implementing create view
func (s *SchemaSuite) TestGetViews() {

}

func (s *SchemaSuite) TestSql() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver.String(), func() {
			schema, mockOrm := GetTestSchema(s.T(), testQuery)
			mockTransaction(mockOrm, testQuery)

			s.NoError(schema.Create("sql", func(table contractsschema.Blueprint) {
				table.String("name")
			}))

			mockOrm.EXPECT().Query().Return(testQuery.Query()).Once()

			schema.Sql("insert into goravel_sql (name) values ('goravel');")

			var count int64
			err := testQuery.Query().Table("sql").Where("name", "goravel").Count(&count)

			s.NoError(err)
			s.Equal(int64(1), count)
		})
	}
}

func mockTransaction(mockOrm *mocksorm.Orm, testQuery *gorm.TestQuery) {
	mockOrm.EXPECT().Transaction(mock.Anything).RunAndReturn(func(txFunc func(contractsorm.Query) error) error {
		return txFunc(testQuery.Query())
	}).Once()
}
