package schema

import (
	"testing"
	"time"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database"
	contractsschema "github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/docker"
	"github.com/goravel/framework/support/env"
)

type SchemaSuite struct {
	suite.Suite
	driverToTestQuery map[database.Driver]*gorm.TestQuery
}

func TestSchemaSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skip test that using Docker")
	}

	suite.Run(t, &SchemaSuite{})
}

func (s *SchemaSuite) SetupTest() {
	postgresDocker := docker.Postgres()
	s.Require().NoError(postgresDocker.Ready())

	postgresQuery := gorm.NewTestQuery(postgresDocker, true)

	sqliteDocker := docker.Sqlite()
	sqliteQuery := gorm.NewTestQuery(sqliteDocker, true)

	mysqlDocker := docker.Mysql()
	s.Require().NoError(mysqlDocker.Ready())

	mysqlQuery := gorm.NewTestQuery(mysqlDocker, true)

	sqlserverDocker := docker.Sqlserver()
	s.Require().NoError(sqlserverDocker.Ready())

	sqlserverQuery := gorm.NewTestQuery(sqlserverDocker, true)

	s.driverToTestQuery = map[database.Driver]*gorm.TestQuery{
		database.DriverPostgres:  postgresQuery,
		database.DriverSqlite:    sqliteQuery,
		database.DriverMysql:     mysqlQuery,
		database.DriverSqlserver: sqlserverQuery,
	}
}

func (s *SchemaSuite) TearDownTest() {
	if s.driverToTestQuery[database.DriverSqlite] != nil {
		s.NoError(s.driverToTestQuery[database.DriverSqlite].Docker().Stop())
	}
}

func (s *SchemaSuite) TestColumnExtraAttributes() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver.String(), func() {
			schema := GetTestSchema(testQuery, s.driverToTestQuery)
			table := "column_extra_attribute"

			s.NoError(schema.Create(table, func(table contractsschema.Blueprint) {
				table.ID()
				table.String("name")
				table.String("nullable").Nullable()
				table.String("string_default").Default("goravel")
				table.Integer("integer_default").Default(1)
				table.Integer("bool_default").Default(true)
				table.TimestampTz("use_current").UseCurrent()
				table.TimestampTz("use_current_on_update").UseCurrent().UseCurrentOnUpdate()
			}))

			type ColumnExtraAttribute struct {
				ID                 uint            `gorm:"primaryKey" json:"id"`
				Name               string          `json:"name"`
				Nullable           *string         `json:"nullable"`
				StringDefault      string          `json:"string_default"`
				IntegerDefault     int             `json:"integer_default"`
				BoolDefault        bool            `json:"bool_default"`
				UseCurrent         carbon.DateTime `json:"use_current"`
				UseCurrentOnUpdate carbon.DateTime `json:"use_current_on_update"`
			}

			// SubSecond: Avoid milliseconds difference
			carbon.SetTimezone(carbon.UTC)
			now := carbon.Now().SubSecond()

			s.NoError(testQuery.Query().Model(&ColumnExtraAttribute{}).Create(map[string]any{
				"name": "hello",
			}))

			interval := int64(1)
			var columnExtraAttribute ColumnExtraAttribute
			s.NoError(testQuery.Query().Where("name", "hello").First(&columnExtraAttribute))
			s.True(columnExtraAttribute.ID > 0)
			s.Equal("hello", columnExtraAttribute.Name)
			s.Nil(columnExtraAttribute.Nullable)
			s.Equal("goravel", columnExtraAttribute.StringDefault)
			s.Equal(1, columnExtraAttribute.IntegerDefault)
			s.True(columnExtraAttribute.BoolDefault)
			s.True(columnExtraAttribute.UseCurrent.Between(now, carbon.Now().AddSecond()))
			s.True(columnExtraAttribute.UseCurrentOnUpdate.Between(now, carbon.Now().AddSecond()))

			time.Sleep(time.Duration(interval) * time.Second)

			now = carbon.Now().SubSecond()
			result, err := testQuery.Query().Model(&ColumnExtraAttribute{}).Where("id", columnExtraAttribute.ID).Update(map[string]any{
				"name": "world",
			})
			s.NoError(err)
			s.Equal(int64(1), result.RowsAffected)

			var anotherColumnExtraAttribute ColumnExtraAttribute
			s.NoError(testQuery.Query().Where("id", columnExtraAttribute.ID).First(&anotherColumnExtraAttribute))
			s.Equal("world", anotherColumnExtraAttribute.Name)
			s.Equal(columnExtraAttribute.UseCurrent, anotherColumnExtraAttribute.UseCurrent)
			if driver == database.DriverMysql {
				s.NotEqual(columnExtraAttribute.UseCurrentOnUpdate, anotherColumnExtraAttribute.UseCurrentOnUpdate)
				s.True(anotherColumnExtraAttribute.UseCurrentOnUpdate.Between(now, carbon.Now().AddSecond()))
			} else {
				s.Equal(columnExtraAttribute.UseCurrentOnUpdate, anotherColumnExtraAttribute.UseCurrentOnUpdate)
			}
		})
	}
}

func (s *SchemaSuite) TestColumnMethods() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver.String(), func() {
			schema := GetTestSchema(testQuery, s.driverToTestQuery)
			table := "column_methods"

			s.NoError(schema.Create(table, func(table contractsschema.Blueprint) {
				table.String("name")
				table.String("age")
				table.String("weight")
			}))
			s.True(schema.HasColumns(table, []string{"name", "age"}))
			s.NoError(schema.Table(table, func(table contractsschema.Blueprint) {
				table.DropColumn("name", "age")
			}))
			s.False(schema.HasColumns(table, []string{"name", "age"}))
		})
	}
}

func (s *SchemaSuite) TestColumnTypes_Postgres() {
	if s.driverToTestQuery[database.DriverPostgres] == nil {
		s.T().Skip("Skip test")
	}

	testQuery := s.driverToTestQuery[database.DriverPostgres]
	schema := GetTestSchema(testQuery, s.driverToTestQuery)
	table := "postgres_columns"
	s.createTableAndAssertColumnsForColumnMethods(schema, table)

	columns, err := schema.GetColumns(table)
	s.Require().Nil(err)

	for _, column := range columns {
		if column.Name == "another_deleted_at" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.True(column.Nullable)
			s.Equal("timestamp(0) with time zone", column.Type)
			s.Equal("timestamptz", column.TypeName)
		}
		if column.Name == "big_integer" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a big_integer column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("bigint", column.Type)
			s.Equal("int8", column.TypeName)
		}
		if column.Name == "char" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a char column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("character(255)", column.Type)
			s.Equal("bpchar", column.TypeName)
		}
		if column.Name == "created_at" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.True(column.Nullable)
			s.Equal("timestamp(2) without time zone", column.Type)
			s.Equal("timestamp", column.TypeName)
		}
		if column.Name == "date" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a date column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("date", column.Type)
			s.Equal("date", column.TypeName)
		}
		if column.Name == "date_time" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a date time column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("timestamp(3) without time zone", column.Type)
			s.Equal("timestamp", column.TypeName)
		}
		if column.Name == "date_time_tz" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a date time with time zone column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("timestamp(3) with time zone", column.Type)
			s.Equal("timestamptz", column.TypeName)
		}
		if column.Name == "decimal" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a decimal column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("numeric(4,1)", column.Type)
			s.Equal("numeric", column.TypeName)
		}
		if column.Name == "deleted_at" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.True(column.Nullable)
			s.Equal("timestamp(0) without time zone", column.Type)
			s.Equal("timestamp", column.TypeName)
		}
		if column.Name == "double" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a double column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("double precision", column.Type)
			s.Equal("float8", column.TypeName)
		}
		if column.Name == "enum" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a enum column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("character varying(255)", column.Type)
			s.Equal("varchar", column.TypeName)
		}
		if column.Name == "float" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a float column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("real", column.Type)
			s.Equal("float4", column.TypeName)
		}
		if column.Name == "id" {
			s.True(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a id column", column.Comment)
			s.Equal("nextval('goravel_postgres_columns_id_seq'::regclass)", column.Default)
			s.False(column.Nullable)
			s.Equal("bigint", column.Type)
			s.Equal("int8", column.TypeName)
		}
		if column.Name == "integer" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a integer column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("integer", column.Type)
			s.Equal("int4", column.TypeName)
		}
		if column.Name == "integer_default" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a integer_default column", column.Comment)
			s.Equal(1, cast.ToInt(column.Default))
			s.False(column.Nullable)
			s.Equal("integer", column.Type)
			s.Equal("int4", column.TypeName)
		}
		if column.Name == "json" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a json column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("json", column.Type)
			s.Equal("json", column.TypeName)
		}
		if column.Name == "jsonb" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a jsonb column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("jsonb", column.Type)
			s.Equal("jsonb", column.TypeName)
		}
		if column.Name == "long_text" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a long_text column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("text", column.Type)
			s.Equal("text", column.TypeName)
		}
		if column.Name == "medium_text" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a medium_text column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("text", column.Type)
			s.Equal("text", column.TypeName)
		}
		if column.Name == "string" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a string column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("character varying(255)", column.Type)
			s.Equal("varchar", column.TypeName)
		}
		if column.Name == "string_default" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a string_default column", column.Comment)
			s.Equal("'goravel'::character varying", column.Default)
			s.False(column.Nullable)
			s.Equal("character varying(255)", column.Type)
			s.Equal("varchar", column.TypeName)
		}
		if column.Name == "text" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a text column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("text", column.Type)
			s.Equal("text", column.TypeName)
		}
		if column.Name == "time" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a time column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("time(2) without time zone", column.Type)
			s.Equal("time", column.TypeName)
		}
		if column.Name == "time_tz" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a time with time zone column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("time(2) with time zone", column.Type)
			s.Equal("timetz", column.TypeName)
		}
		if column.Name == "timestamp" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a timestamp without time zone column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("timestamp(2) without time zone", column.Type)
			s.Equal("timestamp", column.TypeName)
		}
		if column.Name == "timestamp_tz" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a timestamp with time zone column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("timestamp(2) with time zone", column.Type)
			s.Equal("timestamptz", column.TypeName)
		}
		if column.Name == "timestamp_use_current" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a timestamp_use_current column", column.Comment)
			s.Equal("CURRENT_TIMESTAMP", column.Default)
			s.False(column.Nullable)
			s.Equal("timestamp(0) without time zone", column.Type)
			s.Equal("timestamp", column.TypeName)
		}
		if column.Name == "timestamp_use_current_on_update" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a timestamp_use_current_on_update column", column.Comment)
			s.Equal("CURRENT_TIMESTAMP", column.Default)
			s.False(column.Nullable)
			s.Equal("timestamp(0) without time zone", column.Type)
			s.Equal("timestamp", column.TypeName)
		}
		if column.Name == "tiny_text" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a tiny_text column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("character varying(255)", column.Type)
			s.Equal("varchar", column.TypeName)
		}
		if column.Name == "updated_at" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.True(column.Nullable)
			s.Equal("timestamp(2) without time zone", column.Type)
			s.Equal("timestamp", column.TypeName)
		}
		if column.Name == "unsigned_integer" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a unsigned_integer column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("integer", column.Type)
			s.Equal("int4", column.TypeName)
		}
		if column.Name == "unsigned_big_integer" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a unsigned_big_integer column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("bigint", column.Type)
			s.Equal("int8", column.TypeName)
		}
	}
}

func (s *SchemaSuite) TestColumnTypes_Sqlite() {
	if s.driverToTestQuery[database.DriverSqlite] == nil {
		s.T().Skip("Skip test")
	}

	testQuery := s.driverToTestQuery[database.DriverSqlite]
	schema := GetTestSchema(testQuery, s.driverToTestQuery)
	table := "sqlite_columns"
	s.createTableAndAssertColumnsForColumnMethods(schema, table)

	columns, err := schema.GetColumns(table)
	s.Require().Nil(err)

	for _, column := range columns {
		if column.Name == "another_deleted_at" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.True(column.Nullable)
			s.Equal("datetime", column.Type)
		}
		if column.Name == "big_integer" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("integer", column.Type)
		}
		if column.Name == "char" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("varchar", column.Type)
		}
		if column.Name == "created_at" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.True(column.Nullable)
			s.Equal("datetime", column.Type)
		}
		if column.Name == "date" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("date", column.Type)
		}
		if column.Name == "date_time" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("datetime", column.Type)
		}
		if column.Name == "date_time_tz" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("datetime", column.Type)
		}
		if column.Name == "decimal" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("numeric", column.Type)
		}
		if column.Name == "deleted_at" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.True(column.Nullable)
			s.Equal("datetime", column.Type)
		}
		if column.Name == "double" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("double", column.Type)
		}
		if column.Name == "enum" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("varchar", column.Type)
		}
		if column.Name == "float" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("float", column.Type)
		}
		if column.Name == "id" {
			s.True(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("integer", column.Type)
		}
		if column.Name == "integer" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("integer", column.Type)
		}
		if column.Name == "integer_default" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Equal("'1'", column.Default)
			s.False(column.Nullable)
			s.Equal("integer", column.Type)
		}
		if column.Name == "json" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("text", column.Type)
		}
		if column.Name == "jsonb" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("text", column.Type)
		}
		if column.Name == "long_text" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("text", column.Type)
		}
		if column.Name == "medium_text" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("text", column.Type)
		}
		if column.Name == "string" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("varchar", column.Type)
		}
		if column.Name == "string_default" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Equal("'goravel'", column.Default)
			s.False(column.Nullable)
			s.Equal("varchar", column.Type)
		}
		if column.Name == "text" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("text", column.Type)
		}
		if column.Name == "tiny_text" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("text", column.Type)
		}
		if column.Name == "time" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("time", column.Type)
		}
		if column.Name == "time_tz" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("time", column.Type)
		}
		if column.Name == "timestamp" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("datetime", column.Type)
		}
		if column.Name == "timestamp_tz" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("datetime", column.Type)
		}
		if column.Name == "timestamp_use_current" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Equal("CURRENT_TIMESTAMP", column.Default)
			s.False(column.Nullable)
			s.Equal("datetime", column.Type)
		}
		if column.Name == "timestamp_use_current_on_update" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Equal("CURRENT_TIMESTAMP", column.Default)
			s.False(column.Nullable)
			s.Equal("datetime", column.Type)
		}
		if column.Name == "updated_at" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.True(column.Nullable)
			s.Equal("datetime", column.Type)
		}
		if column.Name == "unsigned_integer" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("integer", column.Type)
		}
		if column.Name == "unsigned_big_integer" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("integer", column.Type)
		}
	}
}

func (s *SchemaSuite) TestColumnTypes_Mysql() {
	if s.driverToTestQuery[database.DriverMysql] == nil {
		s.T().Skip("Skip test")
	}

	testQuery := s.driverToTestQuery[database.DriverMysql]
	schema := GetTestSchema(testQuery, s.driverToTestQuery)
	table := "mysql_columns"
	s.createTableAndAssertColumnsForColumnMethods(schema, table)

	columns, err := schema.GetColumns(table)
	s.Require().Nil(err)

	for _, column := range columns {
		if column.Name == "another_deleted_at" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.True(column.Nullable)
			s.Equal("timestamp", column.Type)
			s.Equal("timestamp", column.TypeName)
		}
		if column.Name == "big_integer" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a big_integer column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("bigint", column.Type)
			s.Equal("bigint", column.TypeName)
		}
		if column.Name == "char" {
			s.False(column.Autoincrement)
			s.Equal("utf8mb4_0900_ai_ci", column.Collation)
			s.Equal("This is a char column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("char(255)", column.Type)
			s.Equal("char", column.TypeName)
		}
		if column.Name == "created_at" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.True(column.Nullable)
			s.Equal("timestamp(2)", column.Type)
			s.Equal("timestamp", column.TypeName)
		}
		if column.Name == "date" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a date column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("date", column.Type)
			s.Equal("date", column.TypeName)
		}
		if column.Name == "date_time" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a date time column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("datetime(3)", column.Type)
			s.Equal("datetime", column.TypeName)
		}
		if column.Name == "date_time_tz" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a date time with time zone column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("datetime(3)", column.Type)
			s.Equal("datetime", column.TypeName)
		}
		if column.Name == "decimal" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a decimal column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("decimal(4,1)", column.Type)
			s.Equal("decimal", column.TypeName)
		}
		if column.Name == "deleted_at" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.True(column.Nullable)
			s.Equal("timestamp", column.Type)
			s.Equal("timestamp", column.TypeName)
		}
		if column.Name == "double" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a double column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("double", column.Type)
			s.Equal("double", column.TypeName)
		}
		if column.Name == "enum" {
			s.False(column.Autoincrement)
			s.Equal("utf8mb4_0900_ai_ci", column.Collation)
			s.Equal("This is a enum column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("enum('a','b','c')", column.Type)
			s.Equal("enum", column.TypeName)
		}
		if column.Name == "float" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a float column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("float", column.Type)
			s.Equal("float", column.TypeName)
		}
		if column.Name == "id" {
			s.True(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a id column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("bigint", column.Type)
			s.Equal("bigint", column.TypeName)
		}
		if column.Name == "integer" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a integer column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("int", column.Type)
			s.Equal("int", column.TypeName)
		}
		if column.Name == "integer_default" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a integer_default column", column.Comment)
			s.Equal(1, cast.ToInt(column.Default))
			s.False(column.Nullable)
			s.Equal("int", column.Type)
			s.Equal("int", column.TypeName)
		}
		if column.Name == "json" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a json column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("json", column.Type)
			s.Equal("json", column.TypeName)
		}
		if column.Name == "jsonb" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a jsonb column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("json", column.Type)
			s.Equal("json", column.TypeName)
		}
		if column.Name == "long_text" {
			s.False(column.Autoincrement)
			s.Equal("utf8mb4_0900_ai_ci", column.Collation)
			s.Equal("This is a long_text column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("longtext", column.Type)
			s.Equal("longtext", column.TypeName)
		}
		if column.Name == "medium_text" {
			s.False(column.Autoincrement)
			s.Equal("utf8mb4_0900_ai_ci", column.Collation)
			s.Equal("This is a medium_text column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("mediumtext", column.Type)
			s.Equal("mediumtext", column.TypeName)
		}
		if column.Name == "string" {
			s.False(column.Autoincrement)
			s.Equal("utf8mb4_0900_ai_ci", column.Collation)
			s.Equal("This is a string column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("varchar(255)", column.Type)
			s.Equal("varchar", column.TypeName)
		}
		if column.Name == "string_default" {
			s.False(column.Autoincrement)
			s.Equal("utf8mb4_0900_ai_ci", column.Collation)
			s.Equal("This is a string_default column", column.Comment)
			s.Equal("goravel", column.Default)
			s.False(column.Nullable)
			s.Equal("varchar(255)", column.Type)
			s.Equal("varchar", column.TypeName)
		}
		if column.Name == "text" {
			s.False(column.Autoincrement)
			s.Equal("utf8mb4_0900_ai_ci", column.Collation)
			s.Equal("This is a text column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("text", column.Type)
			s.Equal("text", column.TypeName)
		}
		if column.Name == "time" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a time column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("time(2)", column.Type)
			s.Equal("time", column.TypeName)
		}
		if column.Name == "time_tz" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a time with time zone column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("time(2)", column.Type)
			s.Equal("time", column.TypeName)
		}
		if column.Name == "timestamp" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a timestamp without time zone column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("timestamp(2)", column.Type)
			s.Equal("timestamp", column.TypeName)
		}
		if column.Name == "timestamp_tz" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a timestamp with time zone column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("timestamp(2)", column.Type)
			s.Equal("timestamp", column.TypeName)
		}
		if column.Name == "timestamp_use_current" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a timestamp_use_current column", column.Comment)
			s.Equal("CURRENT_TIMESTAMP", column.Default)
			s.False(column.Nullable)
			s.Equal("timestamp", column.Type)
			s.Equal("timestamp", column.TypeName)
		}
		if column.Name == "timestamp_use_current_on_update" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a timestamp_use_current_on_update column", column.Comment)
			s.Equal("CURRENT_TIMESTAMP", column.Default)
			s.False(column.Nullable)
			s.Equal("timestamp", column.Type)
			s.Equal("timestamp", column.TypeName)
		}
		if column.Name == "tiny_text" {
			s.False(column.Autoincrement)
			s.Equal("utf8mb4_0900_ai_ci", column.Collation)
			s.Equal("This is a tiny_text column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("tinytext", column.Type)
			s.Equal("tinytext", column.TypeName)
		}
		if column.Name == "updated_at" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.True(column.Nullable)
			s.Equal("timestamp(2)", column.Type)
			s.Equal("timestamp", column.TypeName)
		}
		if column.Name == "unsigned_integer" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a unsigned_integer column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("int", column.Type)
			s.Equal("int", column.TypeName)
		}
		if column.Name == "unsigned_big_integer" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a unsigned_big_integer column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("bigint", column.Type)
			s.Equal("bigint", column.TypeName)
		}
	}
}

func (s *SchemaSuite) TestColumnTypes_Sqlserver() {
	if s.driverToTestQuery[database.DriverSqlserver] == nil {
		s.T().Skip("Skip test")
	}

	testQuery := s.driverToTestQuery[database.DriverSqlserver]
	schema := GetTestSchema(testQuery, s.driverToTestQuery)
	table := "sqlserver_columns"
	s.createTableAndAssertColumnsForColumnMethods(schema, table)

	columns, err := schema.GetColumns(table)
	s.Require().Nil(err)

	for _, column := range columns {
		if column.Name == "another_deleted_at" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.True(column.Nullable)
			s.Equal("datetimeoffset(34)", column.Type)
			s.Equal("datetimeoffset", column.TypeName)
		}
		if column.Name == "big_integer" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("bigint", column.Type)
			s.Equal("bigint", column.TypeName)
		}
		if column.Name == "char" {
			s.False(column.Autoincrement)
			s.Equal("SQL_Latin1_General_CP1_CI_AS", column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("nchar(510)", column.Type)
			s.Equal("nchar", column.TypeName)
		}
		if column.Name == "created_at" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.True(column.Nullable)
			s.Equal("datetime2(22)", column.Type)
			s.Equal("datetime2", column.TypeName)
		}
		if column.Name == "date" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("date", column.Type)
			s.Equal("date", column.TypeName)
		}
		if column.Name == "date_time" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("datetime2(23)", column.Type)
			s.Equal("datetime2", column.TypeName)
		}
		if column.Name == "date_time_tz" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("datetimeoffset(30)", column.Type)
			s.Equal("datetimeoffset", column.TypeName)
		}
		if column.Name == "decimal" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("decimal(4,1)", column.Type)
			s.Equal("decimal", column.TypeName)
		}
		if column.Name == "deleted_at" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.True(column.Nullable)
			s.Equal("datetime", column.Type)
			s.Equal("datetime", column.TypeName)
		}
		if column.Name == "double" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("float(53)", column.Type)
			s.Equal("float", column.TypeName)
		}
		if column.Name == "enum" {
			s.False(column.Autoincrement)
			s.Equal("SQL_Latin1_General_CP1_CI_AS", column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("nvarchar(510)", column.Type)
			s.Equal("nvarchar", column.TypeName)
		}
		if column.Name == "float" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("real", column.Type)
			s.Equal("real", column.TypeName)
		}
		if column.Name == "id" {
			s.True(column.Autoincrement)
			s.Empty(column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("bigint", column.Type)
			s.Equal("bigint", column.TypeName)
		}
		if column.Name == "integer" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("int", column.Type)
			s.Equal("int", column.TypeName)
		}
		if column.Name == "integer_default" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Empty(column.Comment)
			s.Equal("('1')", column.Default)
			s.False(column.Nullable)
			s.Equal("int", column.Type)
			s.Equal("int", column.TypeName)
		}
		if column.Name == "json" {
			s.False(column.Autoincrement)
			s.Equal("SQL_Latin1_General_CP1_CI_AS", column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("nvarchar(max)", column.Type)
			s.Equal("nvarchar", column.TypeName)
		}
		if column.Name == "jsonb" {
			s.False(column.Autoincrement)
			s.Equal("SQL_Latin1_General_CP1_CI_AS", column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("nvarchar(max)", column.Type)
			s.Equal("nvarchar", column.TypeName)
		}
		if column.Name == "long_text" {
			s.False(column.Autoincrement)
			s.Equal("SQL_Latin1_General_CP1_CI_AS", column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("nvarchar(max)", column.Type)
			s.Equal("nvarchar", column.TypeName)
		}
		if column.Name == "medium_text" {
			s.False(column.Autoincrement)
			s.Equal("SQL_Latin1_General_CP1_CI_AS", column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("nvarchar(max)", column.Type)
			s.Equal("nvarchar", column.TypeName)
		}
		if column.Name == "string" {
			s.False(column.Autoincrement)
			s.Equal("SQL_Latin1_General_CP1_CI_AS", column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("nvarchar(510)", column.Type)
			s.Equal("nvarchar", column.TypeName)
		}
		if column.Name == "string_default" {
			s.False(column.Autoincrement)
			s.Equal("SQL_Latin1_General_CP1_CI_AS", column.Collation)
			s.Empty(column.Comment)
			s.Equal("('goravel')", column.Default)
			s.False(column.Nullable)
			s.Equal("nvarchar(510)", column.Type)
			s.Equal("nvarchar", column.TypeName)
		}
		if column.Name == "text" {
			s.False(column.Autoincrement)
			s.Equal("SQL_Latin1_General_CP1_CI_AS", column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("nvarchar(max)", column.Type)
			s.Equal("nvarchar", column.TypeName)
		}
		if column.Name == "time" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("time(11)", column.Type)
			s.Equal("time", column.TypeName)
		}
		if column.Name == "time_tz" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("time(11)", column.Type)
			s.Equal("time", column.TypeName)
		}
		if column.Name == "timestamp" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("datetime2(22)", column.Type)
			s.Equal("datetime2", column.TypeName)
		}
		if column.Name == "timestamp_tz" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("datetimeoffset(29)", column.Type)
			s.Equal("datetimeoffset", column.TypeName)
		}
		if column.Name == "timestamp_use_current" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Empty(column.Comment)
			s.Equal("(getdate())", column.Default)
			s.False(column.Nullable)
			s.Equal("datetime", column.Type)
			s.Equal("datetime", column.TypeName)
		}
		if column.Name == "timestamp_use_current_on_update" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Empty(column.Comment)
			s.Equal("(getdate())", column.Default)
			s.False(column.Nullable)
			s.Equal("datetime", column.Type)
			s.Equal("datetime", column.TypeName)
		}
		if column.Name == "tiny_text" {
			s.False(column.Autoincrement)
			s.Equal("SQL_Latin1_General_CP1_CI_AS", column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("nvarchar(510)", column.Type)
			s.Equal("nvarchar", column.TypeName)
		}
		if column.Name == "updated_at" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.True(column.Nullable)
			s.Equal("datetime2(22)", column.Type)
			s.Equal("datetime2", column.TypeName)
		}
		if column.Name == "unsigned_integer" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("int", column.Type)
			s.Equal("int", column.TypeName)
		}
		if column.Name == "unsigned_big_integer" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("bigint", column.Type)
			s.Equal("bigint", column.TypeName)
		}
	}
}

func (s *SchemaSuite) TestTableMethods() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver.String(), func() {
			schema := GetTestSchema(testQuery, s.driverToTestQuery)
			tableOne := "table_one"
			tableTwo := "table_two"
			tableThree := "table_three"

			s.NoError(schema.DropIfExists(tableOne))
			s.NoError(schema.Create(tableOne, func(table contractsschema.Blueprint) {
				table.String("name")
			}))
			s.NoError(schema.Create(tableTwo, func(table contractsschema.Blueprint) {
				table.String("name")
			}))
			s.NoError(schema.Create(tableThree, func(table contractsschema.Blueprint) {
				table.String("name")
			}))
			s.True(schema.HasTable(tableOne))
			s.True(schema.HasTable(tableTwo))
			s.True(schema.HasTable(tableThree))

			tables, err := schema.GetTables()

			s.NoError(err)
			s.Len(tables, 3)

			s.NoError(schema.DropIfExists(tableOne))
			s.False(schema.HasTable(tableOne))

			s.NoError(schema.Table(tableTwo, func(table contractsschema.Blueprint) {
				table.Drop()
			}))
			s.False(schema.HasTable(tableTwo))

			testQuery.MockConfig().EXPECT().GetString("database.connections.postgres.search_path").Return("").Once()

			s.NoError(schema.DropAllTables())
			s.False(schema.HasTable(tableThree))
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
				table.BigInteger("foreign1_id")
				table.Foreign("foreign1_id").References("id").On(table1)
			})

			s.Require().Nil(err)
			s.Require().True(schema.HasTable(table2))

			switch driver {
			case database.DriverPostgres:
				s.True(schema.HasIndex(table2, "goravel_foreign2_pkey"))
			case database.DriverMysql:
				s.True(schema.HasIndex(table2, "goravel_foreign2_foreign1_id_foreign"))
			case database.DriverSqlserver:
				s.True(schema.HasIndex(table2, "goravel_foreign2_foreign1_id_foreign"))
			}
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

			// SQLite does not support set primary index separately
			if driver == database.DriverPostgres {
				s.Require().True(schema.HasIndex(table, "goravel_primaries_pkey"))
			}
			if driver == database.DriverMysql {
				s.Require().True(schema.HasIndex(table, "primary"))
			}
			if driver == database.DriverSqlserver {
				s.Require().True(schema.HasIndex(table, "goravel_primaries_name_age_primary"))
			}
		})
	}
}

func (s *SchemaSuite) TestID_Postgres() {
	if s.driverToTestQuery[database.DriverPostgres] == nil {
		s.T().Skip("Skip test")
	}

	testQuery := s.driverToTestQuery[database.DriverPostgres]
	schema := GetTestSchema(testQuery, s.driverToTestQuery)

	tests := []struct {
		table          string
		setup          func(table string) error
		expectDefault  string
		expectType     string
		expectTypeName string
	}{
		{
			table: "ID",
			setup: func(table string) error {
				return schema.Create(table, func(table contractsschema.Blueprint) {
					table.ID("id").Comment("This is a id column")
				})
			},
			expectDefault:  `nextval('"goravel_ID_id_seq"'::regclass)`,
			expectType:     "bigint",
			expectTypeName: "int8",
		},
		{
			table: "MediumIncrements",
			setup: func(table string) error {
				return schema.Create(table, func(table contractsschema.Blueprint) {
					table.MediumIncrements("id").Comment("This is a id column")
				})
			},
			expectDefault:  `nextval('"goravel_MediumIncrements_id_seq"'::regclass)`,
			expectType:     "integer",
			expectTypeName: "int4",
		},
		{
			table: "IntegerIncrements",
			setup: func(table string) error {
				return schema.Create(table, func(table contractsschema.Blueprint) {
					table.IntegerIncrements("id").Comment("This is a id column")
				})
			},
			expectDefault:  `nextval('"goravel_IntegerIncrements_id_seq"'::regclass)`,
			expectType:     "integer",
			expectTypeName: "int4",
		},
		{
			table: "SmallIncrements",
			setup: func(table string) error {
				return schema.Create(table, func(table contractsschema.Blueprint) {
					table.SmallIncrements("id").Comment("This is a id column")
				})
			},
			expectDefault:  `nextval('"goravel_SmallIncrements_id_seq"'::regclass)`,
			expectType:     "smallint",
			expectTypeName: "int2",
		},
		{
			table: "TinyIncrements",
			setup: func(table string) error {
				return schema.Create(table, func(table contractsschema.Blueprint) {
					table.TinyIncrements("id").Comment("This is a id column")
				})
			},
			expectDefault:  `nextval('"goravel_TinyIncrements_id_seq"'::regclass)`,
			expectType:     "smallint",
			expectTypeName: "int2",
		},
	}

	for _, test := range tests {
		s.Run(test.table, func() {
			s.Require().Nil(test.setup(test.table))
			s.Require().True(schema.HasTable(test.table))

			columns, err := schema.GetColumns(test.table)
			s.Require().Nil(err)
			s.Equal(1, len(columns))
			s.Equal("id", columns[0].Name)
			s.True(columns[0].Autoincrement)
			s.Empty(columns[0].Collation)
			s.Equal("This is a id column", columns[0].Comment)
			s.Equal(test.expectDefault, columns[0].Default)
			s.False(columns[0].Nullable)
			s.Equal(test.expectType, columns[0].Type)
			s.Equal(test.expectTypeName, columns[0].TypeName)
		})
	}
}

func (s *SchemaSuite) TestID_Sqlite() {
	if s.driverToTestQuery[database.DriverSqlite] == nil {
		s.T().Skip("Skip test")
	}

	testQuery := s.driverToTestQuery[database.DriverSqlite]
	schema := GetTestSchema(testQuery, s.driverToTestQuery)

	tests := []struct {
		table      string
		setup      func(table string) error
		expectType string
	}{
		{
			table: "ID",
			setup: func(table string) error {
				return schema.Create(table, func(table contractsschema.Blueprint) {
					table.ID("id").Comment("This is a id column")
				})
			},
			expectType: "integer",
		},
		{
			table: "MediumIncrements",
			setup: func(table string) error {
				return schema.Create(table, func(table contractsschema.Blueprint) {
					table.MediumIncrements("id").Comment("This is a id column")
				})
			},
			expectType: "integer",
		},
		{
			table: "IntegerIncrements",
			setup: func(table string) error {
				return schema.Create(table, func(table contractsschema.Blueprint) {
					table.IntegerIncrements("id").Comment("This is a id column")
				})
			},
			expectType: "integer",
		},
		{
			table: "SmallIncrements",
			setup: func(table string) error {
				return schema.Create(table, func(table contractsschema.Blueprint) {
					table.SmallIncrements("id").Comment("This is a id column")
				})
			},
			expectType: "integer",
		},
		{
			table: "TinyIncrements",
			setup: func(table string) error {
				return schema.Create(table, func(table contractsschema.Blueprint) {
					table.TinyIncrements("id").Comment("This is a id column")
				})
			},
			expectType: "integer",
		},
	}

	for _, test := range tests {
		s.Run(test.table, func() {
			s.Require().Nil(test.setup(test.table))
			s.Require().True(schema.HasTable(test.table))

			columns, err := schema.GetColumns(test.table)
			s.Require().Nil(err)
			s.Equal(1, len(columns))
			s.Equal("id", columns[0].Name)
			s.True(columns[0].Autoincrement)
			s.Empty(columns[0].Comment)
			s.Empty(columns[0].Default)
			s.False(columns[0].Nullable)
			s.Equal(test.expectType, columns[0].Type)
		})
	}
}

func (s *SchemaSuite) TestID_Mysql() {
	if s.driverToTestQuery[database.DriverMysql] == nil {
		s.T().Skip("Skip test")
	}

	testQuery := s.driverToTestQuery[database.DriverMysql]
	schema := GetTestSchema(testQuery, s.driverToTestQuery)

	tests := []struct {
		table          string
		setup          func(table string) error
		expectType     string
		expectTypeName string
	}{
		{
			table: "ID",
			setup: func(table string) error {
				return schema.Create(table, func(table contractsschema.Blueprint) {
					table.ID("id").Comment("This is a id column")
				})
			},
			expectType:     "bigint",
			expectTypeName: "bigint",
		},
		{
			table: "MediumIncrements",
			setup: func(table string) error {
				return schema.Create(table, func(table contractsschema.Blueprint) {
					table.MediumIncrements("id").Comment("This is a id column")
				})
			},
			expectType:     "mediumint",
			expectTypeName: "mediumint",
		},
		{
			table: "IntegerIncrements",
			setup: func(table string) error {
				return schema.Create(table, func(table contractsschema.Blueprint) {
					table.IntegerIncrements("id").Comment("This is a id column")
				})
			},
			expectType:     "int",
			expectTypeName: "int",
		},
		{
			table: "SmallIncrements",
			setup: func(table string) error {
				return schema.Create(table, func(table contractsschema.Blueprint) {
					table.SmallIncrements("id").Comment("This is a id column")
				})
			},
			expectType:     "smallint",
			expectTypeName: "smallint",
		},
		{
			table: "TinyIncrements",
			setup: func(table string) error {
				return schema.Create(table, func(table contractsschema.Blueprint) {
					table.TinyIncrements("id").Comment("This is a id column")
				})
			},
			expectType:     "tinyint",
			expectTypeName: "tinyint",
		},
	}

	for _, test := range tests {
		s.Run(test.table, func() {
			s.Require().Nil(test.setup(test.table))
			s.Require().True(schema.HasTable(test.table))

			columns, err := schema.GetColumns(test.table)
			s.Require().Nil(err)
			s.Equal(1, len(columns))
			s.True(columns[0].Autoincrement)
			s.Empty(columns[0].Collation)
			s.Equal("This is a id column", columns[0].Comment)
			s.Empty(columns[0].Default)
			s.False(columns[0].Nullable)
			s.Equal(test.expectType, columns[0].Type)
			s.Equal(test.expectTypeName, columns[0].TypeName)
		})
	}
}

func (s *SchemaSuite) TestID_Sqlserver() {
	if s.driverToTestQuery[database.DriverSqlserver] == nil {
		s.T().Skip("Skip test")
	}

	testQuery := s.driverToTestQuery[database.DriverSqlserver]
	schema := GetTestSchema(testQuery, s.driverToTestQuery)

	tests := []struct {
		table          string
		setup          func(table string) error
		expectType     string
		expectTypeName string
	}{
		{
			table: "ID",
			setup: func(table string) error {
				return schema.Create(table, func(table contractsschema.Blueprint) {
					table.ID("id").Comment("This is a id column")
				})
			},
			expectType:     "bigint",
			expectTypeName: "bigint",
		},
		{
			table: "MediumIncrements",
			setup: func(table string) error {
				return schema.Create(table, func(table contractsschema.Blueprint) {
					table.MediumIncrements("id").Comment("This is a id column")
				})
			},
			expectType:     "int",
			expectTypeName: "int",
		},
		{
			table: "IntegerIncrements",
			setup: func(table string) error {
				return schema.Create(table, func(table contractsschema.Blueprint) {
					table.IntegerIncrements("id").Comment("This is a id column")
				})
			},
			expectType:     "int",
			expectTypeName: "int",
		},
		{
			table: "SmallIncrements",
			setup: func(table string) error {
				return schema.Create(table, func(table contractsschema.Blueprint) {
					table.SmallIncrements("id").Comment("This is a id column")
				})
			},
			expectType:     "smallint",
			expectTypeName: "smallint",
		},
		{
			table: "TinyIncrements",
			setup: func(table string) error {
				return schema.Create(table, func(table contractsschema.Blueprint) {
					table.TinyIncrements("id").Comment("This is a id column")
				})
			},
			expectType:     "tinyint",
			expectTypeName: "tinyint",
		},
	}

	for _, test := range tests {
		s.Run(test.table, func() {
			s.Require().Nil(test.setup(test.table))
			s.Require().True(schema.HasTable(test.table))

			columns, err := schema.GetColumns(test.table)
			s.Require().Nil(err)
			s.Equal(1, len(columns))
			s.True(columns[0].Autoincrement)
			s.Empty(columns[0].Collation)
			s.Empty(columns[0].Comment)
			s.Empty(columns[0].Default)
			s.False(columns[0].Nullable)
			s.Equal(test.expectType, columns[0].Type)
			s.Equal(test.expectTypeName, columns[0].TypeName)
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
					} else if driver == database.DriverSqlserver {
						s.Equal("nonclustered", index.Type)
					} else {
						s.Equal("btree", index.Type)
					}
					s.False(index.Unique)
				}
			}
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

func (s *SchemaSuite) createTableAndAssertColumnsForColumnMethods(schema contractsschema.Schema, table string) {
	err := schema.Create(table, func(table contractsschema.Blueprint) {
		table.BigInteger("big_integer").Comment("This is a big_integer column")
		table.Char("char").Comment("This is a char column")
		table.Date("date").Comment("This is a date column")
		table.DateTime("date_time", 3).Comment("This is a date time column")
		table.DateTimeTz("date_time_tz", 3).Comment("This is a date time with time zone column")
		table.Decimal("decimal").Places(1).Total(4).Comment("This is a decimal column")
		table.Double("double").Comment("This is a double column")
		table.Enum("enum", []string{"a", "b", "c"}).Comment("This is a enum column")
		table.Float("float", 2).Comment("This is a float column")
		table.LongText("long_text").Comment("This is a long_text column")
		table.MediumText("medium_text").Comment("This is a medium_text column")
		table.ID().Comment("This is a id column")
		table.Integer("integer").Comment("This is a integer column")
		table.Integer("integer_default").Default(1).Comment("This is a integer_default column")
		table.Json("json").Comment("This is a json column")
		table.Jsonb("jsonb").Comment("This is a jsonb column")
		table.SoftDeletes()
		table.SoftDeletesTz("another_deleted_at")
		table.String("string").Comment("This is a string column")
		table.String("string_default").Default("goravel").Comment("This is a string_default column")
		table.Text("text").Comment("This is a text column")
		table.TinyText("tiny_text").Comment("This is a tiny_text column")
		table.Time("time", 2).Comment("This is a time column")
		table.TimeTz("time_tz", 2).Comment("This is a time with time zone column")
		table.Timestamp("timestamp", 2).Comment("This is a timestamp without time zone column")
		table.TimestampTz("timestamp_tz", 2).Comment("This is a timestamp with time zone column")
		table.Timestamps(2)
		table.Timestamp("timestamp_use_current").UseCurrent().Comment("This is a timestamp_use_current column")
		table.Timestamp("timestamp_use_current_on_update").UseCurrent().UseCurrentOnUpdate().Comment("This is a timestamp_use_current_on_update column")
		table.UnsignedInteger("unsigned_integer").Comment("This is a unsigned_integer column")
		table.UnsignedBigInteger("unsigned_big_integer").Comment("This is a unsigned_big_integer column")
	})

	s.Require().Nil(err)
	s.Require().True(schema.HasTable(table))
	s.True(schema.HasColumn(table, "big_integer"))
	s.True(schema.HasColumns(table, []string{"big_integer", "decimal"}))

	columnListing := schema.GetColumnListing(table)

	s.Equal(32, len(columnListing))
	s.Contains(columnListing, "another_deleted_at")
	s.Contains(columnListing, "big_integer")
	s.Contains(columnListing, "char")
	s.Contains(columnListing, "created_at")
	s.Contains(columnListing, "date")
	s.Contains(columnListing, "date_time")
	s.Contains(columnListing, "date_time_tz")
	s.Contains(columnListing, "decimal")
	s.Contains(columnListing, "deleted_at")
	s.Contains(columnListing, "double")
	s.Contains(columnListing, "enum")
	s.Contains(columnListing, "float")
	s.Contains(columnListing, "id")
	s.Contains(columnListing, "integer")
	s.Contains(columnListing, "integer_default")
	s.Contains(columnListing, "json")
	s.Contains(columnListing, "jsonb")
	s.Contains(columnListing, "long_text")
	s.Contains(columnListing, "medium_text")
	s.Contains(columnListing, "string")
	s.Contains(columnListing, "string_default")
	s.Contains(columnListing, "text")
	s.Contains(columnListing, "tiny_text")
	s.Contains(columnListing, "time")
	s.Contains(columnListing, "time_tz")
	s.Contains(columnListing, "timestamp")
	s.Contains(columnListing, "timestamp_tz")
	s.Contains(columnListing, "timestamp_use_current")
	s.Contains(columnListing, "timestamp_use_current_on_update")
	s.Contains(columnListing, "unsigned_integer")
	s.Contains(columnListing, "unsigned_big_integer")
	s.Contains(columnListing, "updated_at")
}
