package schema

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database"
	contractsschema "github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/support/docker"
	"github.com/goravel/framework/support/env"
)

type SchemaSuite struct {
	suite.Suite
	driverToTestQuery map[database.Driver]*gorm.TestQuery
}

func TestSchemaSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests that use Docker")
	}

	suite.Run(t, &SchemaSuite{})
}

func (s *SchemaSuite) SetupTest() {
	// TODO Add other drivers
	//postgresDocker := docker.Postgres()
	//postgresQuery := gorm.NewTestQuery(postgresDocker, true)
	//
	//sqliteDocker := docker.Sqlite()
	//sqliteQuery := gorm.NewTestQuery(sqliteDocker, true)

	mysqlDocker := docker.Mysql()
	mysqlQuery := gorm.NewTestQuery(mysqlDocker, true)

	s.driverToTestQuery = map[database.Driver]*gorm.TestQuery{
		//database.DriverPostgres: postgresQuery,
		//database.DriverSqlite:   sqliteQuery,
		database.DriverMysql: mysqlQuery,
	}
}

func (s *SchemaSuite) TestCreate_DropIfExists_HasTable() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver.String(), func() {
			schema := GetTestSchema(testQuery, s.driverToTestQuery)
			table := "drop_if_exists"

			s.NoError(schema.DropIfExists(table))
			s.NoError(schema.Create(table, func(table contractsschema.Blueprint) {
				table.String("name")
			}))
			s.True(schema.HasTable(table))
			s.NoError(schema.DropIfExists(table))
			s.False(schema.HasTable(table))
		})
	}
}

func (s *SchemaSuite) TestDropAllTables() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver.String(), func() {
			schema := GetTestSchema(testQuery, s.driverToTestQuery)
			tableOne := "drop_all1_tables"
			tableTwo := "drop_all2_tables"

			s.NoError(schema.Create(tableOne, func(table contractsschema.Blueprint) {
				table.String("name")
			}))

			s.NoError(schema.Create(tableTwo, func(table contractsschema.Blueprint) {
				table.String("name")
			}))
			s.True(schema.HasTable(tableOne))
			s.True(schema.HasTable(tableTwo))

			testQuery.MockConfig().EXPECT().GetString("database.connections.postgres.search_path").Return("").Once()

			s.NoError(schema.DropAllTables())
			s.False(schema.HasTable(tableOne))
			s.False(schema.HasTable(tableTwo))
		})
	}
}

// TODO Implement this after implementing create type
func (s *SchemaSuite) TestDropAllTypes() {

}

// TODO Implement this after implementing create view
func (s *SchemaSuite) TestDropAllViews() {

}

func (s *SchemaSuite) TestForeign() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver.String(), func() {
			schema := GetTestSchema(testQuery, s.driverToTestQuery)
			table1 := "foreign1"

			err := schema.Create(table1, func(table contractsschema.Blueprint) {
				table.ID()
				table.String("name")
			})

			s.Require().Nil(err)
			s.Require().True(schema.HasTable(table1))

			table2 := "foreign2"
			err = schema.Create(table2, func(table contractsschema.Blueprint) {
				table.ID()
				table.String("name")
				table.Integer("foreign1_id")
				table.Foreign("foreign1_id").References("id").On(table1)
			})

			s.Require().Nil(err)
			s.Require().True(schema.HasTable(table2))
		})
	}
}

func (s *SchemaSuite) TestPrimary() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver.String(), func() {
			schema := GetTestSchema(testQuery, s.driverToTestQuery)
			table := "primaries"

			s.NoError(schema.Create(table, func(table contractsschema.Blueprint) {
				table.String("name")
				table.String("age")
				table.Primary("name", "age")
			}))

			s.Require().True(schema.HasTable(table))
			if driver != database.DriverSqlite {
				// SQLite does not support set primary index separately
				s.Require().True(schema.HasIndex(table, "goravel_primaries_pkey"))
			}
		})
	}
}

func (s *SchemaSuite) TestIndexMethods() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver.String(), func() {
			schema := GetTestSchema(testQuery, s.driverToTestQuery)
			table := "indexes"
			err := schema.Create(table, func(table contractsschema.Blueprint) {
				table.ID()
				table.String("name")
				table.Index("id", "name")
				table.Index("name").Name("name_index")
			})

			s.Require().Nil(err)
			s.True(schema.HasTable(table))
			s.Contains(schema.GetIndexListing(table), "goravel_indexes_id_name_index")
			s.True(schema.HasIndex(table, "goravel_indexes_id_name_index"))
			s.True(schema.HasIndex(table, "name_index"))

			indexes, err := schema.GetIndexes(table)
			s.Require().Nil(err)
			s.Len(indexes, 3)

			for _, index := range indexes {
				if index.Name == "goravel_indexes_id_name_index" {
					s.ElementsMatch(index.Columns, []string{"id", "name"})
					s.False(index.Primary)
					if driver == database.DriverSqlite {
						s.Empty(index.Type)
					} else {
						s.Equal("btree", index.Type)
					}
					s.False(index.Unique)
				}
			}
		})
	}
}

func (s *SchemaSuite) TestTable_GetTables() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver.String(), func() {
			schema := GetTestSchema(testQuery, s.driverToTestQuery)

			s.NoError(schema.Create("changes", func(table contractsschema.Blueprint) {
				table.String("name")
			}))
			s.True(schema.HasTable("changes"))

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

func (s *SchemaSuite) TestSql() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver.String(), func() {
			schema := GetTestSchema(testQuery, s.driverToTestQuery)

			s.NoError(schema.Create("sql", func(table contractsschema.Blueprint) {
				table.String("name")
			}))

			schema.Sql("insert into goravel_sql (name) values ('goravel');")

			var count int64
			err := testQuery.Query().Table("sql").Where("name", "goravel").Count(&count)

			s.NoError(err)
			s.Equal(int64(1), count)
		})
	}
}
