package schema

import (
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	ormcontract "github.com/goravel/framework/contracts/database/orm"
	schemacontract "github.com/goravel/framework/contracts/database/schema"
	testingcontract "github.com/goravel/framework/contracts/testing"
	"github.com/goravel/framework/database/gorm"
	configmock "github.com/goravel/framework/mocks/config"
	ormmock "github.com/goravel/framework/mocks/database/orm"
	logmock "github.com/goravel/framework/mocks/log"
	"github.com/goravel/framework/support/env"
)

type TestSchema struct {
	dbConfig   testingcontract.DatabaseConfig
	driver     ormcontract.Driver
	mockConfig *configmock.Config
	mockOrm    *ormmock.Orm
	mockLog    *logmock.Log
	query      ormcontract.Query
	schema     *Schema
}

type SchemaSuite struct {
	suite.Suite
	schemas []TestSchema
	//mysqlQuery      ormcontract.Query
	postgresQuery ormcontract.Query
	//sqliteQuery     ormcontract.Query
	//sqlserverDB     ormcontract.Query
}

func TestSchemaSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	if err := testDatabaseDocker.Fresh(); err != nil {
		t.Fatal(err)
	}

	//mysqlDocker := gorm.NewMysqlDocker(testDatabaseDocker)
	//mysqlQuery, err := mysqlDocker.New()
	//if err != nil {
	//	log.Fatalf("Init mysql docker error: %v", err)
	//}

	postgresqlDocker := gorm.NewPostgresqlDocker(testDatabaseDocker)
	postgresqlQuery, err := postgresqlDocker.New()
	if err != nil {
		log.Fatalf("Init postgresql docker error: %v", err)
	}

	//sqliteDocker := gorm.NewSqliteDocker("goravel")
	//sqliteQuery, err := sqliteDocker.New()
	//if err != nil {
	//	log.Fatalf("Get sqlite error: %s", err)
	//}
	//
	//sqlserverDocker := gorm.NewSqlserverDocker(testDatabaseDocker)
	//sqlserverQuery, err := sqlserverDocker.New()
	//if err != nil {
	//	log.Fatalf("Init sqlserver docker error: %v", err)
	//}

	suite.Run(t, &SchemaSuite{
		//mysqlQuery:      mysqlQuery,
		postgresQuery: postgresqlQuery,
		//sqliteQuery:     sqliteQuery,
		//sqlserverDB:     sqlserverQuery,
	})

	//assert.Nil(t, file.Remove("goravel"))
}

func (s *SchemaSuite) SetupTest() {
	dbConfig := testDatabaseDocker.Postgres.Config()
	mockConfig := &configmock.Config{}
	mockOrm := &ormmock.Orm{}
	//mockOrmOfConnection := &ormmock.Orm{}
	mockLog := &logmock.Log{}
	mockConfig.On("GetString", "database.default").Return("mysql").Once()
	mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.database", "mysql")).
		Return(dbConfig.Database).Once()
	mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.schema", "mysql")).
		Return("").Once()
	mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.prefix", "mysql")).
		Return("").Once()
	mockOrm.On("Connection", "mysql").Return(mockOrm).Once()
	mockOrm.On("Query").Return(s.postgresQuery).Once()
	postgresSchema, err := NewSchema("", mockConfig, mockOrm, mockLog)
	s.Nil(err)

	s.schemas = []TestSchema{
		{
			dbConfig:   testDatabaseDocker.Postgres.Config(),
			driver:     ormcontract.DriverPostgres,
			mockConfig: mockConfig,
			mockOrm:    mockOrm,
			mockLog:    mockLog,
			query:      s.postgresQuery,
			schema:     postgresSchema,
		},
	}
}

func (s *SchemaSuite) TestConnection() {
	for _, schema := range s.schemas {
		schema.mockOrm.On("Connection", "postgres").Return(schema.mockOrm).Once()
		schema.mockOrm.On("Query").Return(schema.query).Once()
		schema.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.database", "postgres")).
			Return(schema.dbConfig.Database).Once()
		schema.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.schema", "postgres")).
			Return("").Once()
		schema.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.prefix", "postgres")).
			Return("").Once()
		s.NotNil(schema.schema.Connection("postgres"))

		schema.mockConfig.AssertExpectations(s.T())
		schema.mockOrm.AssertExpectations(s.T())
	}
}

func (s *SchemaSuite) TestCreate() {
	for _, schema := range s.schemas {
		s.Run(schema.driver.String(), func() {
			err := schema.schema.Create("creates", func(table schemacontract.Blueprint) {
				table.String("name")
				table.Comment("This is a test table")
			})
			s.Nil(err)
			s.True(schema.schema.HasTable("creates"))

			schema.mockConfig.AssertExpectations(s.T())
		})
	}
}

func (s *SchemaSuite) TestDrop() {
	for _, schema := range s.schemas {
		s.Run(schema.driver.String(), func() {
			table := "drops"
			err := schema.schema.Create(table, func(table schemacontract.Blueprint) {
				table.String("name")
			})
			s.Nil(err)
			s.True(schema.schema.HasTable(table))

			err = schema.schema.Drop(table)

			s.Nil(err)
			s.False(schema.schema.HasTable(table))

			schema.mockConfig.AssertExpectations(s.T())
		})
	}
}

func (s *SchemaSuite) TestDropAllTables() {
	for _, schema := range s.schemas {
		s.Run(schema.driver.String(), func() {
			table1 := "drop_all_table1"
			err := schema.schema.Create(table1, func(table schemacontract.Blueprint) {
				table.ID()
			})
			s.Nil(err)
			s.True(schema.schema.HasTable(table1))

			table2 := "drop_all_table2"
			err = schema.schema.Create(table2, func(table schemacontract.Blueprint) {
				table.ID()
			})
			s.Nil(err)
			s.True(schema.schema.HasTable(table2))

			err = schema.schema.DropAllTables()
			s.Nil(err)
			s.False(schema.schema.HasTable(table1))
			s.False(schema.schema.HasTable(table2))

			schema.mockConfig.AssertExpectations(s.T())
		})
	}
}

func (s *SchemaSuite) TestDropColumns() {
	for _, schema := range s.schemas {
		s.Run(schema.driver.String(), func() {
			table := "drop_columns"
			err := schema.schema.Create(table, func(table schemacontract.Blueprint) {
				table.String("name")
				table.String("summary")
				table.Comment("This is a test table")
			})
			s.Nil(err)
			s.True(schema.schema.HasTable(table))
			s.True(schema.schema.HasColumns(table, []string{"name", "summary"}))

			err = schema.schema.DropColumns(table, []string{"summary"})

			s.Nil(err)
			s.True(schema.schema.HasColumn(table, "name"))
			s.False(schema.schema.HasColumn(table, "summary"))

			schema.mockConfig.AssertExpectations(s.T())
		})
	}
}

func (s *SchemaSuite) TestDropForeign() {
	for _, schema := range s.schemas {
		s.Run(schema.driver.String(), func() {
			table1 := "drop_foreign1"
			err := schema.schema.Create(table1, func(table schemacontract.Blueprint) {
				table.ID()
				table.String("name")
			})

			s.Require().Nil(err)
			s.Require().True(schema.schema.HasTable(table1))

			table2 := "drop_foreign2"
			err = schema.schema.Create(table2, func(table schemacontract.Blueprint) {
				table.ID()
				table.String("name")
				table.UnsignedInteger("drop_foreign1_id")
				table.Foreign([]string{"drop_foreign1_id"}).References("id").On(table1)
			})

			s.Require().Nil(err)
			s.Require().True(schema.schema.HasTable(table2))

			err = schema.schema.Table(table2, func(table schemacontract.Blueprint) {
				table.DropForeign([]string{"drop_foreign1_id"})
			})

			s.Require().Nil(err)

			schema.mockConfig.AssertExpectations(s.T())
		})
	}
}

func (s *SchemaSuite) TestDropIfExists() {
	for _, schema := range s.schemas {
		s.Run(schema.driver.String(), func() {
			table := "drop_if_exists"

			err := schema.schema.DropIfExists(table)
			s.Nil(err)

			err = schema.schema.Create(table, func(table schemacontract.Blueprint) {
				table.String("name")
			})
			s.Nil(err)
			s.True(schema.schema.HasTable(table))

			err = schema.schema.DropIfExists(table)
			s.Nil(err)

			s.False(schema.schema.HasTable(table))

			schema.mockConfig.AssertExpectations(s.T())
		})
	}
}

func (s *SchemaSuite) TestDropIndex() {
	for _, schema := range s.schemas {
		s.Run(schema.driver.String(), func() {
			table := "drop_indexes"
			err := schema.schema.Create(table, func(table schemacontract.Blueprint) {
				table.ID()
				table.String("name")
				table.Index([]string{"id", "name"})
			})

			s.Require().Nil(err)
			s.Require().True(schema.schema.HasTable(table))
			s.Require().True(schema.schema.HasIndex(table, "drop_indexes_id_name_index"))

			err = schema.schema.Table(table, func(table schemacontract.Blueprint) {
				table.DropIndex([]string{"id", "name"})
			})
			s.Require().Nil(err)
			s.Require().False(schema.schema.HasIndex(table, "drop_indexes_id_name_index"))

			schema.mockConfig.AssertExpectations(s.T())
		})
	}
}

func (s *SchemaSuite) TestDropIndexByName() {
	for _, schema := range s.schemas {
		s.Run(schema.driver.String(), func() {
			table := "drop_index_by_name"
			err := schema.schema.Create(table, func(table schemacontract.Blueprint) {
				table.ID()
				table.String("name")
				table.Index([]string{"id", "name"})
			})

			s.Require().Nil(err)
			s.Require().True(schema.schema.HasTable(table))
			s.Require().True(schema.schema.HasIndex(table, "drop_index_by_name_id_name_index"))

			err = schema.schema.Table(table, func(table schemacontract.Blueprint) {
				table.DropIndexByName("drop_index_by_name_id_name_index")
			})
			s.Require().Nil(err)
			s.Require().False(schema.schema.HasIndex(table, "drop_index_by_name_id_name_index"))

			schema.mockConfig.AssertExpectations(s.T())
		})
	}
}

func (s *SchemaSuite) TestDropSoftDeletes() {
	for _, schema := range s.schemas {
		s.Run(schema.driver.String(), func() {
			table := "drop_soft_deletes"
			err := schema.schema.Create(table, func(table schemacontract.Blueprint) {
				table.SoftDeletes()
			})

			s.Require().Nil(err)
			s.Require().True(schema.schema.HasTable(table))
			s.Require().True(schema.schema.HasColumn(table, "deleted_at"))

			err = schema.schema.Table(table, func(table schemacontract.Blueprint) {
				table.DropSoftDeletes()
			})
			s.Require().Nil(err)
			s.Require().False(schema.schema.HasColumn(table, "deleted_at"))

			schema.mockConfig.AssertExpectations(s.T())
		})
	}
}

func (s *SchemaSuite) TestDropTimestamps() {
	for _, schema := range s.schemas {
		s.Run(schema.driver.String(), func() {
			table := "drop_timestamps"
			err := schema.schema.Create(table, func(table schemacontract.Blueprint) {
				table.Timestamps()
			})

			s.Require().Nil(err)
			s.Require().True(schema.schema.HasTable(table))
			s.Require().True(schema.schema.HasColumn(table, "created_at"))
			s.Require().True(schema.schema.HasColumn(table, "updated_at"))

			err = schema.schema.Table(table, func(table schemacontract.Blueprint) {
				table.DropTimestamps()
			})
			s.Require().Nil(err)
			s.Require().False(schema.schema.HasColumn(table, "created_at"))
			s.Require().False(schema.schema.HasColumn(table, "updated_at"))

			schema.mockConfig.AssertExpectations(s.T())
		})
	}
}

func (s *SchemaSuite) TestGetColumns() {
	for _, schema := range s.schemas {
		s.Run(schema.driver.String(), func() {
			table := "get_columns"
			err := schema.schema.Create(table, func(table schemacontract.Blueprint) {
				table.BigIncrements("big_increments").Comment("This is a big_increments column")
				table.BigInteger("big_integer").Comment("This is a big_integer column")
				table.Char("char").Comment("This is a char column")
				table.Date("date").Comment("This is a date column")
				table.DateTime("date_time", 3).Comment("This is a date time column")
				table.DateTimeTz("date_time_tz", 3).Comment("This is a date time with time zone column")
				table.Decimal("decimal", schemacontract.DecimalConfig{Places: 1, Total: 4}).Comment("This is a decimal column")
				table.Double("double").Comment("This is a double column")
				table.Enum("enum", []string{"a", "b", "c"}).Comment("This is a enum column")
				table.Float("float", 2).Comment("This is a float column")
				table.ID().Comment("This is a id column")
				table.ID("aid").Comment("This is a id column, name is aid")
				table.Integer("integer").Comment("This is a integer column")
				table.SoftDeletes()
				table.SoftDeletesTz("another_deleted_at")
				table.String("string").Comment("This is a string column")
				table.Json("json").Comment("This is a json column")
				table.Jsonb("jsonb").Comment("This is a jsonb column")
				table.Text("text").Comment("This is a text column")
				table.Time("time", 2).Comment("This is a time column")
				table.TimeTz("time_tz", 2).Comment("This is a time with time zone column")
				table.Timestamp("timestamp", 2).Comment("This is a timestamp without time zone column")
				table.TimestampTz("timestamp_tz", 2).Comment("This is a timestamp with time zone column")
				table.Timestamps(2)
				table.UnsignedInteger("unsigned_integer").Comment("This is a unsigned_integer column")
				table.UnsignedBigInteger("unsigned_big_integer").Comment("This is a unsigned_big_integer column")
			})

			s.Require().Nil(err)
			s.Require().True(schema.schema.HasTable(table))

			columnListing := schema.schema.GetColumnListing(table)

			s.Equal(27, len(columnListing))
			s.Contains(columnListing, "big_increments")
			s.Contains(columnListing, "big_integer")
			s.Contains(columnListing, "char")
			s.Contains(columnListing, "date")
			s.Contains(columnListing, "date_time")
			s.Contains(columnListing, "date_time_tz")
			s.Contains(columnListing, "decimal")
			s.Contains(columnListing, "double")
			s.Contains(columnListing, "enum")
			s.Contains(columnListing, "id")
			s.Contains(columnListing, "aid")
			s.Contains(columnListing, "integer")
			s.Contains(columnListing, "deleted_at")
			s.Contains(columnListing, "another_deleted_at")
			s.Contains(columnListing, "string")
			s.Contains(columnListing, "json")
			s.Contains(columnListing, "jsonb")
			s.Contains(columnListing, "text")
			s.Contains(columnListing, "time")
			s.Contains(columnListing, "time_tz")
			s.Contains(columnListing, "timestamp")
			s.Contains(columnListing, "timestamp_tz")
			s.Contains(columnListing, "created_at")
			s.Contains(columnListing, "updated_at")
			s.Contains(columnListing, "unsigned_integer")
			s.Contains(columnListing, "unsigned_big_integer")

			columns, err := schema.schema.GetColumns(table)
			s.Require().Nil(err)
			for _, column := range columns {
				if column.Name == "big_increments" {
					s.True(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Equal("This is a big_increments column", column.Comment)
					s.Equal("nextval('get_columns_big_increments_seq'::regclass)", column.Default)
					s.False(column.Nullable)
					s.Equal("bigint", column.Type)
					s.Equal("int8", column.TypeName)
				}
				if column.Name == "big_integer" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Equal("This is a big_integer column", column.Comment)
					s.Empty(column.Default)
					s.False(column.Nullable)
					s.Equal("bigint", column.Type)
					s.Equal("int8", column.TypeName)
				}
				if column.Name == "char" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Equal("This is a char column", column.Comment)
					s.Empty(column.Default)
					s.False(column.Nullable)
					s.Equal("character(255)", column.Type)
					s.Equal("bpchar", column.TypeName)
				}
				if column.Name == "date" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Equal("This is a date column", column.Comment)
					s.Empty(column.Default)
					s.False(column.Nullable)
					s.Equal("date", column.Type)
					s.Equal("date", column.TypeName)
				}
				if column.Name == "date_time" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Equal("This is a date time column", column.Comment)
					s.Empty(column.Default)
					s.False(column.Nullable)
					s.Equal("timestamp(3) without time zone", column.Type)
					s.Equal("timestamp", column.TypeName)
				}
				if column.Name == "date_time_tz" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Equal("This is a date time with time zone column", column.Comment)
					s.Empty(column.Default)
					s.False(column.Nullable)
					s.Equal("timestamp(3) with time zone", column.Type)
					s.Equal("timestamptz", column.TypeName)
				}
				if column.Name == "decimal" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Equal("This is a decimal column", column.Comment)
					s.Empty(column.Default)
					s.False(column.Nullable)
					s.Equal("numeric(4,1)", column.Type)
					s.Equal("numeric", column.TypeName)
				}
				if column.Name == "double" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Equal("This is a double column", column.Comment)
					s.Empty(column.Default)
					s.False(column.Nullable)
					s.Equal("double precision", column.Type)
					s.Equal("float8", column.TypeName)
				}
				if column.Name == "enum" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Equal("This is a enum column", column.Comment)
					s.Empty(column.Default)
					s.False(column.Nullable)
					s.Equal("character varying(255)", column.Type)
					s.Equal("varchar", column.TypeName)
				}
				if column.Name == "float" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Equal("This is a float column", column.Comment)
					s.Empty(column.Default)
					s.False(column.Nullable)
					s.Equal("real", column.Type)
					s.Equal("float4", column.TypeName)
				}
				if column.Name == "id" {
					s.True(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Equal("This is a id column", column.Comment)
					s.Equal("nextval('get_columns_id_seq'::regclass)", column.Default)
					s.False(column.Nullable)
					s.Equal("bigint", column.Type)
					s.Equal("int8", column.TypeName)
				}
				if column.Name == "aid" {
					s.True(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Equal("This is a id column, name is aid", column.Comment)
					s.Equal("nextval('get_columns_aid_seq'::regclass)", column.Default)
					s.False(column.Nullable)
					s.Equal("bigint", column.Type)
					s.Equal("int8", column.TypeName)
				}
				if column.Name == "integer" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Equal("This is a integer column", column.Comment)
					s.Empty(column.Default)
					s.False(column.Nullable)
					s.Equal("integer", column.Type)
					s.Equal("int4", column.TypeName)
				}
				if column.Name == "deleted_at" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Empty(column.Comment)
					s.Empty(column.Default)
					s.True(column.Nullable)
					s.Equal("timestamp(0) without time zone", column.Type)
					s.Equal("timestamp", column.TypeName)
				}
				if column.Name == "another_deleted_at" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Empty(column.Comment)
					s.Empty(column.Default)
					s.True(column.Nullable)
					s.Equal("timestamp(0) with time zone", column.Type)
					s.Equal("timestamptz", column.TypeName)
				}
				if column.Name == "string" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Equal("This is a string column", column.Comment)
					s.Empty(column.Default)
					s.False(column.Nullable)
					s.Equal("character varying(255)", column.Type)
					s.Equal("varchar", column.TypeName)
				}
				if column.Name == "json" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Equal("This is a json column", column.Comment)
					s.Empty(column.Default)
					s.False(column.Nullable)
					s.Equal("json", column.Type)
					s.Equal("json", column.TypeName)
				}
				if column.Name == "jsonb" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Equal("This is a jsonb column", column.Comment)
					s.Empty(column.Default)
					s.False(column.Nullable)
					s.Equal("jsonb", column.Type)
					s.Equal("jsonb", column.TypeName)
				}
				if column.Name == "text" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Equal("This is a text column", column.Comment)
					s.Empty(column.Default)
					s.False(column.Nullable)
					s.Equal("text", column.Type)
					s.Equal("text", column.TypeName)
				}
				if column.Name == "time" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Equal("This is a time column", column.Comment)
					s.Empty(column.Default)
					s.False(column.Nullable)
					s.Equal("time(2) without time zone", column.Type)
					s.Equal("time", column.TypeName)
				}
				if column.Name == "time_tz" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Equal("This is a time with time zone column", column.Comment)
					s.Empty(column.Default)
					s.False(column.Nullable)
					s.Equal("time(2) with time zone", column.Type)
					s.Equal("timetz", column.TypeName)
				}
				if column.Name == "timestamp" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Equal("This is a timestamp without time zone column", column.Comment)
					s.Empty(column.Default)
					s.False(column.Nullable)
					s.Equal("timestamp(2) without time zone", column.Type)
					s.Equal("timestamp", column.TypeName)
				}
				if column.Name == "timestamp_tz" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Equal("This is a timestamp with time zone column", column.Comment)
					s.Empty(column.Default)
					s.False(column.Nullable)
					s.Equal("timestamp(2) with time zone", column.Type)
					s.Equal("timestamptz", column.TypeName)
				}
				if column.Name == "created_at" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Empty(column.Comment)
					s.Empty(column.Default)
					s.True(column.Nullable)
					s.Equal("timestamp(2) without time zone", column.Type)
					s.Equal("timestamp", column.TypeName)
				}
				if column.Name == "updated_at" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Empty(column.Comment)
					s.Empty(column.Default)
					s.True(column.Nullable)
					s.Equal("timestamp(2) without time zone", column.Type)
					s.Equal("timestamp", column.TypeName)
				}
				if column.Name == "unsigned_integer" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Equal("This is a unsigned_integer column", column.Comment)
					s.Empty(column.Default)
					s.False(column.Nullable)
					s.Equal("integer", column.Type)
					s.Equal("int4", column.TypeName)
				}
				if column.Name == "unsigned_big_integer" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Equal("This is a unsigned_big_integer column", column.Comment)
					s.Empty(column.Default)
					s.False(column.Nullable)
					s.Equal("bigint", column.Type)
					s.Equal("int8", column.TypeName)
				}
			}

			s.True(schema.schema.HasColumn(table, "char"))
			s.True(schema.schema.HasColumns(table, []string{"char", "string"}))

			schema.mockConfig.AssertExpectations(s.T())
		})
	}
}

func (s *SchemaSuite) TestGetIndexes() {
	for _, schema := range s.schemas {
		s.Run(schema.driver.String(), func() {
			table := "get_indexes"
			err := schema.schema.Create(table, func(table schemacontract.Blueprint) {
				table.ID()
				table.String("name")
				table.Index([]string{"id", "name"})
			})

			s.Require().Nil(err)
			s.Require().True(schema.schema.HasTable(table))
			s.Require().Contains(schema.schema.GetIndexListing(table), "get_indexes_id_name_index")
			s.Require().True(schema.schema.HasIndex(table, "get_indexes_id_name_index"))

			indexes, err := schema.schema.GetIndexes(table)
			s.Require().Nil(err)
			s.Len(indexes, 1)
			for _, index := range indexes {
				if index.Name == "get_indexes_id_name_index" {
					s.ElementsMatch(index.Columns, []string{"id", "name"})
					s.False(index.Primary)
					s.Equal("btree", index.Type)
					s.False(index.Unique)
				}
			}

			schema.mockConfig.AssertExpectations(s.T())
		})
	}
}

func (s *SchemaSuite) TestGetTables() {
	for _, schema := range s.schemas {
		s.Run(schema.driver.String(), func() {
			tables, err := schema.schema.GetTables()
			s.Greater(len(tables), 0)
			s.Nil(err)
		})
	}
}

func (s *SchemaSuite) TestGetTableListing() {
	for _, schema := range s.schemas {
		s.Run(schema.driver.String(), func() {
			tables := schema.schema.GetTableListing()
			s.Greater(len(tables), 0)
		})
	}
}

func (s *SchemaSuite) TestForeign() {
	for _, schema := range s.schemas {
		s.Run(schema.driver.String(), func() {
			table1 := "foreign1"
			err := schema.schema.Create(table1, func(table schemacontract.Blueprint) {
				table.ID()
				table.String("name")
			})

			s.Require().Nil(err)
			s.Require().True(schema.schema.HasTable(table1))

			table2 := "foreign2"
			err = schema.schema.Create(table2, func(table schemacontract.Blueprint) {
				table.ID()
				table.String("name")
				table.UnsignedInteger("foreign1_id")
				table.Foreign([]string{"foreign1_id"}).References("id").On(table1)
			})

			s.Require().Nil(err)
			s.Require().True(schema.schema.HasTable(table2))

			schema.mockConfig.AssertExpectations(s.T())
		})
	}
}

func (s *SchemaSuite) TestHasTable() {
	for _, schema := range s.schemas {
		s.Run(schema.driver.String(), func() {
			s.True(schema.schema.HasTable("users"))
			s.False(schema.schema.HasTable("unknow"))
		})
	}
}

func (s *SchemaSuite) TestInitGrammarAndProcess() {
	for _, schema := range s.schemas {
		s.Run(schema.driver.String(), func() {
			s.Nil(schema.schema.initGrammarAndProcess())
			grammarType := reflect.TypeOf(schema.schema.grammar)
			grammarName := grammarType.Elem().Name()
			processorType := reflect.TypeOf(schema.schema.processor)
			processorName := processorType.Elem().Name()

			switch schema.driver {
			case ormcontract.DriverMysql:
				s.Equal("Mysql", grammarName)
				s.Equal("Mysql", processorName)
			case ormcontract.DriverPostgres:
				s.Equal("Postgres", grammarName)
				s.Equal("Postgres", processorName)
			case ormcontract.DriverSqlserver:
				s.Equal("Sqlserver", grammarName)
				s.Equal("Sqlserver", processorName)
			case ormcontract.DriverSqlite:
				s.Equal("Sqlite", grammarName)
				s.Equal("Sqlite", processorName)
			default:
				s.Fail("unsupported database driver")
			}
		})
	}
}

func (s *SchemaSuite) TestParseDatabaseAndSchemaAndTable() {
	for _, schema := range s.schemas {
		s.Run(schema.driver.String(), func() {
			database, schemaName, table := schema.schema.parseDatabaseAndSchemaAndTable("users")
			s.Equal(schema.dbConfig.Database, database)
			s.Equal("public", schemaName)
			s.Equal("users", table)

			database, schemaName, table = schema.schema.parseDatabaseAndSchemaAndTable("goravel.users")
			s.Equal(schema.dbConfig.Database, database)
			s.Equal("goravel", schemaName)
			s.Equal("users", table)

			schema.mockConfig.AssertExpectations(s.T())
		})
	}
}

func (s *SchemaSuite) TestPrimary() {
	for _, schema := range s.schemas {
		s.Run(schema.driver.String(), func() {
			table := "primaries"
			err := schema.schema.Create(table, func(table schemacontract.Blueprint) {
				table.String("name")
				table.String("age")
				table.Primary([]string{"name", "age"})
			})

			s.Require().Nil(err)
			s.Require().True(schema.schema.HasTable(table))
			s.Require().True(schema.schema.HasIndex(table, "primaries_pkey"))

			schema.mockConfig.AssertExpectations(s.T())
		})
	}
}

func (s *SchemaSuite) TestRename() {
	for _, schema := range s.schemas {
		s.Run(schema.driver.String(), func() {
			table := "renames"
			err := schema.schema.Create(table, func(table schemacontract.Blueprint) {
				table.String("name")
			})
			s.Require().Nil(err)
			s.Require().True(schema.schema.HasTable(table))

			table1 := "renames1"
			err = schema.schema.Rename(table, table1)
			s.Require().Nil(err)
			s.Require().False(schema.schema.HasTable(table))
			s.Require().True(schema.schema.HasTable(table1))

			schema.mockConfig.AssertExpectations(s.T())
		})
	}
}

func (s *SchemaSuite) TestRenameColumn() {
	for _, schema := range s.schemas {
		s.Run(schema.driver.String(), func() {
			table := "rename_column"
			err := schema.schema.Create(table, func(table schemacontract.Blueprint) {
				table.String("name")
			})
			s.Require().Nil(err)
			s.Require().True(schema.schema.HasTable(table))
			s.Require().True(schema.schema.HasColumn(table, "name"))

			err = schema.schema.Table(table, func(table schemacontract.Blueprint) {
				table.RenameColumn("name", "age")
			})
			s.Require().Nil(err)
			s.Require().False(schema.schema.HasColumn(table, "name"))
			s.Require().True(schema.schema.HasColumn(table, "age"))

			schema.mockConfig.AssertExpectations(s.T())
		})
	}
}

func (s *SchemaSuite) TestRenameIndex() {
	for _, schema := range s.schemas {
		s.Run(schema.driver.String(), func() {
			table := "rename_index"
			err := schema.schema.Create(table, func(table schemacontract.Blueprint) {
				table.ID()
				table.String("name")
				table.Index([]string{"id", "name"})
			})
			s.Require().Nil(err)
			s.Require().True(schema.schema.HasTable(table))
			s.Require().True(schema.schema.HasIndex(table, "rename_index_id_name_index"))

			err = schema.schema.Table(table, func(table schemacontract.Blueprint) {
				table.RenameIndex("rename_index_id_name_index", "rename_index_id_name_index1")
			})
			s.Require().Nil(err)
			s.Require().False(schema.schema.HasIndex(table, "rename_index_id_name_index"))
			s.Require().True(schema.schema.HasIndex(table, "rename_index_id_name_index1"))

			schema.mockConfig.AssertExpectations(s.T())
		})
	}
}

func (s *SchemaSuite) TestUnique() {
	for _, schema := range s.schemas {
		s.Run(schema.driver.String(), func() {
			table := "uniques"
			err := schema.schema.Create(table, func(table schemacontract.Blueprint) {
				table.String("name")
				table.String("age")
				table.Unique([]string{"name", "age"})
			})

			s.Require().Nil(err)
			s.Require().True(schema.schema.HasTable(table))
			s.Require().True(schema.schema.HasIndex(table, "uniques_name_age_unique"))

			schema.mockConfig.AssertExpectations(s.T())
		})
	}
}

//func (s *SchemaSuite) TestTable() {
//	for _, schema := range s.schemas {
//		s.Run(schema.driver.String(), func() {
//			schema.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.prefix", schema.schema.connection)).
//				Return("goravel_").Twice()
//
//			err := schema.schema.Create("table", func(table schemacontract.Blueprint) {
//				table.String("name")
//			})
//			s.Nil(err)
//			s.True(schema.schema.HasTable("goravel_table"))
//
//			columns, err := schema.schema.GetColumns("goravel_table")
//			s.Require().Nil(err)
//			for _, column := range columns {
//				if column.Name == "name" {
//					s.False(column.AutoIncrement)
//					s.Empty(column.Collation)
//					s.Empty(column.Comment)
//					s.Empty(column.Default)
//					s.False(column.Nullable)
//					s.Equal("character varying(255)", column.Type)
//					s.Equal("varchar", column.TypeName)
//				}
//			}
//
//			err = schema.schema.Table("table", func(table schemacontract.Blueprint) {
//				table.String("name").Comment("This is a name column").Change()
//			})
//			s.Nil(err)
//			s.True(schema.schema.HasTable("goravel_table"))
//
//			schema.mockConfig.AssertExpectations(s.T())
//		})
//	}
//}
