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
	mockConfig := &configmock.Config{}
	mockOrm := &ormmock.Orm{}
	//mockOrmOfConnection := &ormmock.Orm{}
	mockLog := &logmock.Log{}
	mockConfig.On("GetString", "database.default").Return("mysql").Once()
	mockOrm.On("Connection", "mysql").Return(mockOrm).Once()
	mockOrm.On("Query").Return(s.postgresQuery).Once()
	postgresSchema, err := NewSchema("", mockConfig, mockOrm, mockLog)
	s.Nil(err)
	s.schemas = append(s.schemas, TestSchema{
		dbConfig:   testDatabaseDocker.Postgres.Config(),
		driver:     ormcontract.DriverPostgres,
		mockConfig: mockConfig,
		mockOrm:    mockOrm,
		mockLog:    mockLog,
		query:      s.postgresQuery,
		schema:     postgresSchema,
	})
}

func (s *SchemaSuite) TestConnection() {
	for _, schema := range s.schemas {
		schema.mockOrm.On("Connection", "postgres").Return(schema.mockOrm).Once()
		schema.mockOrm.On("Query").Return(schema.query).Once()
		s.NotNil(schema.schema.Connection("postgres"))

		schema.mockOrm.AssertExpectations(s.T())
	}
}

func (s *SchemaSuite) TestCreate() {
	for _, schema := range s.schemas {
		s.Run(schema.driver.String(), func() {
			schema.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.prefix", schema.schema.connection)).
				Return("goravel_").Once()

			err := schema.schema.Create("create", func(table schemacontract.Blueprint) {
				table.String("name")
				table.Comment("This is a test table")
			})
			s.Nil(err)
			s.True(schema.schema.HasTable("goravel_create"))

			schema.mockConfig.AssertExpectations(s.T())
		})
	}
}

func (s *SchemaSuite) TestDropColumns() {
	for _, schema := range s.schemas {
		s.Run(schema.driver.String(), func() {
			schema.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.database", schema.schema.connection)).
				Return(schema.dbConfig.Database).Times(3)
			schema.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.schema", schema.schema.connection)).
				Return("").Times(3)
			schema.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.prefix", schema.schema.connection)).
				Return("").Times(5)

			err := schema.schema.Create("drop_columns", func(table schemacontract.Blueprint) {
				table.String("name")
				table.String("summary")
				table.Comment("This is a test table")
			})
			s.Nil(err)
			s.True(schema.schema.HasTable("drop_columns"))
			s.True(schema.schema.HasColumns("drop_columns", []string{"name", "summary"}))

			err = schema.schema.DropColumns("drop_columns", []string{"summary"})

			s.Nil(err)
			s.True(schema.schema.HasColumn("drop_columns", "name"))
			s.False(schema.schema.HasColumn("drop_columns", "summary"))

			schema.mockConfig.AssertExpectations(s.T())
		})
	}
}

func (s *SchemaSuite) TestGetColumns() {
	for _, schema := range s.schemas {
		s.Run(schema.driver.String(), func() {
			table := "get_columns"
			schema.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.database", schema.schema.connection)).
				Return(schema.dbConfig.Database).Times(4)
			schema.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.schema", schema.schema.connection)).
				Return("").Times(4)
			schema.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.prefix", schema.schema.connection)).
				Return("").Times(5)

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
				table.String("string").Comment("This is a string column")
				table.UnsignedBigInteger("unsigned_big_integer").Comment("This is a unsigned_big_integer column")
			})

			s.Require().Nil(err)
			s.Require().True(schema.schema.HasTable(table))

			columnListing := schema.schema.GetColumnListing(table)

			s.Equal(15, len(columnListing))
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
			s.Contains(columnListing, "string")
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
					s.True(column.Nullable)
					s.Equal("bigint", column.Type)
					s.Equal("int8", column.TypeName)
				}
				if column.Name == "char" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Equal("This is a char column", column.Comment)
					s.Empty(column.Default)
					s.True(column.Nullable)
					s.Equal("character(255)", column.Type)
					s.Equal("bpchar", column.TypeName)
				}
				if column.Name == "date" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Equal("This is a date column", column.Comment)
					s.Empty(column.Default)
					s.True(column.Nullable)
					s.Equal("date", column.Type)
					s.Equal("date", column.TypeName)
				}
				if column.Name == "date_time" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Equal("This is a date time column", column.Comment)
					s.Empty(column.Default)
					s.True(column.Nullable)
					s.Equal("timestamp(3) without time zone", column.Type)
					s.Equal("timestamp", column.TypeName)
				}
				if column.Name == "date_time_tz" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Equal("This is a date time with time zone column", column.Comment)
					s.Empty(column.Default)
					s.True(column.Nullable)
					s.Equal("timestamp(3) with time zone", column.Type)
					s.Equal("timestamptz", column.TypeName)
				}
				if column.Name == "decimal" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Equal("This is a decimal column", column.Comment)
					s.Empty(column.Default)
					s.True(column.Nullable)
					s.Equal("numeric(4,1)", column.Type)
					s.Equal("numeric", column.TypeName)
				}
				if column.Name == "double" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Equal("This is a double column", column.Comment)
					s.Empty(column.Default)
					s.True(column.Nullable)
					s.Equal("double precision", column.Type)
					s.Equal("float8", column.TypeName)
				}
				if column.Name == "enum" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Equal("This is a enum column", column.Comment)
					s.Empty(column.Default)
					s.True(column.Nullable)
					s.Equal("character varying(255)", column.Type)
					s.Equal("varchar", column.TypeName)
				}
				if column.Name == "float" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Equal("This is a float column", column.Comment)
					s.Empty(column.Default)
					s.True(column.Nullable)
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
					s.True(column.Nullable)
					s.Equal("integer", column.Type)
					s.Equal("int4", column.TypeName)
				}
				if column.Name == "string" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Equal("This is a string column", column.Comment)
					s.Empty(column.Default)
					s.True(column.Nullable)
					s.Equal("character varying(255)", column.Type)
					s.Equal("varchar", column.TypeName)
				}
				if column.Name == "unsigned_big_integer" {
					s.False(column.AutoIncrement)
					s.Empty(column.Collation)
					s.Equal("This is a unsigned_big_integer column", column.Comment)
					s.Empty(column.Default)
					s.True(column.Nullable)
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

func (s *SchemaSuite) TestGetTables() {
	for _, schema := range s.schemas {
		s.Run(schema.driver.String(), func() {
			tables, err := schema.schema.GetTables()
			s.Greater(len(tables), 0)
			s.Nil(err)
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
			schema.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.database", schema.schema.connection)).
				Return(schema.dbConfig.Database).Once()
			schema.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.schema", schema.schema.connection)).
				Return("").Once()
			database, schemaName, table := schema.schema.parseDatabaseAndSchemaAndTable("users")
			s.Equal(schema.dbConfig.Database, database)
			s.Equal("public", schemaName)
			s.Equal("users", table)

			schema.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.database", schema.schema.connection)).
				Return(schema.dbConfig.Database).Once()
			schema.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.schema", schema.schema.connection)).
				Return("").Once()
			database, schemaName, table = schema.schema.parseDatabaseAndSchemaAndTable("goravel.users")
			s.Equal(schema.dbConfig.Database, database)
			s.Equal("goravel", schemaName)
			s.Equal("users", table)

			schema.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.database", schema.schema.connection)).
				Return(schema.dbConfig.Database).Once()
			schema.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.schema", schema.schema.connection)).
				Return("hello").Once()
			database, schemaName, table = schema.schema.parseDatabaseAndSchemaAndTable("goravel.users")
			s.Equal(schema.dbConfig.Database, database)
			s.Equal("goravel", schemaName)
			s.Equal("users", table)

			schema.mockConfig.AssertExpectations(s.T())
		})
	}
}
