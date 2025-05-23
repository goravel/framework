package tests

import (
	"fmt"
	"strings"
	"testing"
	"time"

	contractsschema "github.com/goravel/framework/contracts/database/schema"
	databaseschema "github.com/goravel/framework/database/schema"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/mysql"
	"github.com/goravel/postgres"
	"github.com/goravel/sqlite"
	"github.com/goravel/sqlserver"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SchemaSuite struct {
	suite.Suite
	prefix            string
	driverToTestQuery map[string]*TestQuery
}

func TestSchemaSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &SchemaSuite{
		driverToTestQuery: make(map[string]*TestQuery),
	})
}

func (s *SchemaSuite) SetupTest() {
	s.prefix = "goravel_"
	s.driverToTestQuery = NewTestQueryBuilder().All(s.prefix, true)
}

func (s *SchemaSuite) TearDownTest() {
	if s.driverToTestQuery[sqlite.Name] != nil {
		docker, err := s.driverToTestQuery[sqlite.Name].Driver().Docker()
		s.NoError(err)
		s.NoError(docker.Shutdown())
	}
}

func (s *SchemaSuite) TestColumnChange() {
	for driver, testQuery := range s.driverToTestQuery {
		if driver == sqlite.Name {
			continue
		}
		s.Run(driver, func() {
			schema := newSchema(testQuery, s.driverToTestQuery)
			table := "column_change"
			expectedDefaultStringLength := databaseschema.DefaultStringLength
			customStringLength := 100
			expectedCustomStringLength := customStringLength
			expectedColumnType := "text"

			if driver == sqlserver.Name {
				expectedDefaultStringLength = databaseschema.DefaultStringLength * 2
				expectedCustomStringLength = customStringLength * 2
				expectedColumnType = "nvarchar"
			}
			s.NoError(schema.Create(table, func(table contractsschema.Blueprint) {
				table.ID()
				table.String("change_length")
				table.String("change_type")
				table.String("change_to_nullable")
				table.String("change_to_not_nullable").Nullable()
				table.String("change_add_default")
				table.String("change_remove_default").Default("goravel")
				table.String("change_modify_default").Default("goravel")
				table.String("change_add_comment")
				table.String("change_remove_comment").Comment("goravel")
				table.String("change_modify_comment").Comment("goravel")

			}))
			columns, err := schema.GetColumns(table)
			s.Require().Nil(err)
			for _, column := range columns {
				if column.Name == "change_length" {
					s.Contains(column.Type, fmt.Sprintf("(%d)", expectedDefaultStringLength))
				}
				if column.Name == "change_type" {
					s.Contains(column.TypeName, "varchar")
				}
				if column.Name == "change_to_nullable" {
					s.False(column.Nullable)
				}
				if column.Name == "change_to_not_nullable" {
					s.True(column.Nullable)
				}
				if column.Name == "change_add_default" {
					s.Empty(column.Default)
				}
				if column.Name == "change_remove_default" || column.Name == "change_modify_default" {
					s.Contains(column.Default, "goravel")
				}
				if driver != sqlserver.Name {
					if column.Name == "change_add_comment" {
						s.Empty(column.Comment)
					}
					if column.Name == "change_remove_comment" || column.Name == "change_modify_comment" {
						s.Contains(column.Comment, "goravel")
					}
				}

			}
			s.NoError(schema.Table(table, func(table contractsschema.Blueprint) {
				table.String("change_length", customStringLength).Change()
				table.Text("change_type").Change()
				table.String("change_to_nullable").Nullable().Change()
				table.String("change_to_not_nullable").Change()
				table.String("change_add_default").Default("goravel").Change()
				table.String("change_remove_default").Change()
				table.String("change_modify_default").Default("goravel_again").Change()
				table.String("change_add_comment").Comment("goravel").After("change_type").Change()
				table.String("change_remove_comment").Change()
				table.String("change_modify_comment").Comment("goravel_again").First().Change()
			}))
			columns, err = schema.GetColumns(table)
			s.Require().Nil(err)
			for i, column := range columns {
				if driver == mysql.Name {
					if i == 0 {
						s.Equal(column.Name, "change_modify_comment")
					}
					if column.Name == "change_type" {
						s.Equal(columns[i+1].Name, "change_add_comment")
					}
				}
				if column.Name == "change_length" {
					s.Contains(column.Type, fmt.Sprintf("(%d)", expectedCustomStringLength))
				}
				if column.Name == "change_type" {
					s.Contains(column.TypeName, expectedColumnType)
				}
				if column.Name == "change_to_nullable" {
					s.True(column.Nullable)
				}
				if column.Name == "change_to_not_nullable" {
					s.False(column.Nullable)
				}
				if column.Name == "change_add_default" {
					s.Contains(column.Default, "goravel")
				}
				if column.Name == "change_remove_default" {
					s.Empty(column.Default)
				}
				if column.Name == "change_modify_default" {
					s.Contains(column.Default, "goravel_again")
				}
				if driver != sqlserver.Name {
					if column.Name == "change_add_comment" {
						s.Contains(column.Comment, "goravel")
					}
					if column.Name == "change_remove_comment" {
						s.Empty(column.Comment)
					}
					if column.Name == "change_modify_comment" {
						s.Contains(column.Comment, "goravel_again")
					}
				}
			}
		})
	}
}

func (s *SchemaSuite) TestColumnExtraAttributes() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver, func() {
			schema := newSchema(testQuery, s.driverToTestQuery)
			table := "column_extra_attribute"

			s.NoError(schema.Create(table, func(table contractsschema.Blueprint) {
				table.ID()
				table.String("name")
				table.String("nullable").Nullable()
				table.String("string_default").Default("goravel")
				table.Integer("integer_default").Default(1)
				table.Boolean("bool_default").Default(true)
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
			if driver == mysql.Name {
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
		s.Run(driver, func() {
			schema := newSchema(testQuery, s.driverToTestQuery)
			table := "column_methods"

			s.NoError(schema.Create(table, func(table contractsschema.Blueprint) {
				table.String("name")
				table.String("age")
				table.String("weight")
				table.String("height")
			}))
			s.True(schema.HasColumns(table, []string{"name", "age", "weight"}))
			s.NoError(schema.Table(table, func(table contractsschema.Blueprint) {
				table.DropColumn("name", "age")
			}))
			s.NoError(schema.DropColumns(table, []string{"weight"}))
			s.False(schema.HasColumns(table, []string{"name", "age", "weight"}))
		})
	}
}

func (s *SchemaSuite) TestColumnTypes_Postgres() {
	if s.driverToTestQuery[postgres.Name] == nil {
		s.T().Skip("Skip test")
	}

	testQuery := s.driverToTestQuery[postgres.Name]
	schema := newSchema(testQuery, s.driverToTestQuery)
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
		if column.Name == "boolean_default" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a boolean column with default value", column.Comment)
			s.Equal("true", column.Default)
			s.False(column.Nullable)
			s.Equal("boolean", column.Type)
			s.Equal("bool", column.TypeName)
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
		if column.Name == "custom_type" {
			s.False(column.Autoincrement)
			s.Equal("This is a custom type column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("macaddr", column.Type)
			s.Equal("macaddr", column.TypeName)
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
		if column.Name == "enum" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a enum column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("character varying(255)", column.Type)
			s.Equal("varchar", column.TypeName)
		}
		if column.Name == "enum_int" {
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
	if s.driverToTestQuery[sqlite.Name] == nil {
		s.T().Skip("Skip test")
	}

	testQuery := s.driverToTestQuery[sqlite.Name]
	schema := newSchema(testQuery, s.driverToTestQuery)
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
			s.Equal("datetime", column.TypeName)
		}
		if column.Name == "big_integer" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("integer", column.Type)
			s.Equal("integer", column.TypeName)
		}
		if column.Name == "boolean_default" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Equal("'1'", column.Default)
			s.False(column.Nullable)
			s.Equal("tinyint(1)", column.Type)
			s.Equal("tinyint", column.TypeName)
		}
		if column.Name == "char" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("varchar", column.Type)
			s.Equal("varchar", column.TypeName)
		}
		if column.Name == "created_at" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.True(column.Nullable)
			s.Equal("datetime", column.Type)
			s.Equal("datetime", column.TypeName)
		}
		if column.Name == "custom_type" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("geometry", column.Type)
			s.Equal("geometry", column.TypeName)
		}
		if column.Name == "date" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("date", column.Type)
			s.Equal("date", column.TypeName)
		}
		if column.Name == "date_time" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("datetime", column.Type)
			s.Equal("datetime", column.TypeName)
		}
		if column.Name == "date_time_tz" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("datetime", column.Type)
			s.Equal("datetime", column.TypeName)
		}
		if column.Name == "decimal" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("numeric", column.Type)
			s.Equal("numeric", column.TypeName)
		}
		if column.Name == "deleted_at" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.True(column.Nullable)
			s.Equal("datetime", column.Type)
			s.Equal("datetime", column.TypeName)
		}
		if column.Name == "double" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("double", column.Type)
			s.Equal("double", column.TypeName)
		}
		if column.Name == "enum" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("varchar", column.Type)
			s.Equal("varchar", column.TypeName)
		}
		if column.Name == "enum_int" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("varchar", column.Type)
			s.Equal("varchar", column.TypeName)
		}
		if column.Name == "float" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("float", column.Type)
			s.Equal("float", column.TypeName)
		}
		if column.Name == "id" {
			s.True(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("integer", column.Type)
			s.Equal("integer", column.TypeName)
		}
		if column.Name == "integer" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("integer", column.Type)
			s.Equal("integer", column.TypeName)
		}
		if column.Name == "integer_default" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Equal("'1'", column.Default)
			s.False(column.Nullable)
			s.Equal("integer", column.Type)
			s.Equal("integer", column.TypeName)
		}
		if column.Name == "json" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("text", column.Type)
			s.Equal("text", column.TypeName)
		}
		if column.Name == "jsonb" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("text", column.Type)
			s.Equal("text", column.TypeName)
		}
		if column.Name == "long_text" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("text", column.Type)
			s.Equal("text", column.TypeName)
		}
		if column.Name == "medium_text" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("text", column.Type)
			s.Equal("text", column.TypeName)
		}
		if column.Name == "string" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("varchar", column.Type)
			s.Equal("varchar", column.TypeName)
		}
		if column.Name == "string_default" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Equal("'goravel'", column.Default)
			s.False(column.Nullable)
			s.Equal("varchar", column.Type)
			s.Equal("varchar", column.TypeName)
		}
		if column.Name == "text" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("text", column.Type)
			s.Equal("text", column.TypeName)
		}
		if column.Name == "tiny_text" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("text", column.Type)
			s.Equal("text", column.TypeName)
		}
		if column.Name == "time" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("time", column.Type)
			s.Equal("time", column.TypeName)
		}
		if column.Name == "time_tz" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("time", column.Type)
			s.Equal("time", column.TypeName)
		}
		if column.Name == "timestamp" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("datetime", column.Type)
			s.Equal("datetime", column.TypeName)
		}
		if column.Name == "timestamp_tz" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("datetime", column.Type)
			s.Equal("datetime", column.TypeName)
		}
		if column.Name == "timestamp_use_current" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Equal("CURRENT_TIMESTAMP", column.Default)
			s.False(column.Nullable)
			s.Equal("datetime", column.Type)
			s.Equal("datetime", column.TypeName)
		}
		if column.Name == "timestamp_use_current_on_update" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Equal("CURRENT_TIMESTAMP", column.Default)
			s.False(column.Nullable)
			s.Equal("datetime", column.Type)
			s.Equal("datetime", column.TypeName)
		}
		if column.Name == "updated_at" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.True(column.Nullable)
			s.Equal("datetime", column.Type)
			s.Equal("datetime", column.TypeName)
		}
		if column.Name == "unsigned_integer" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("integer", column.Type)
			s.Equal("integer", column.TypeName)
		}
		if column.Name == "unsigned_big_integer" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("integer", column.Type)
			s.Equal("integer", column.TypeName)
		}
	}
}

func (s *SchemaSuite) TestColumnTypes_Mysql() {
	if s.driverToTestQuery[mysql.Name] == nil {
		s.T().Skip("Skip test")
	}

	testQuery := s.driverToTestQuery[mysql.Name]
	schema := newSchema(testQuery, s.driverToTestQuery)
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
		if column.Name == "boolean_default" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a boolean column with default value", column.Comment)
			s.Equal("1", column.Default)
			s.False(column.Nullable)
			s.Equal("tinyint(1)", column.Type)
			s.Equal("tinyint", column.TypeName)
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
		if column.Name == "custom_type" {
			s.False(column.Autoincrement)
			s.Equal("This is a custom type column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("geometry", column.Type)
			s.Equal("geometry", column.TypeName)
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
		if column.Name == "enum_int" {
			s.False(column.Autoincrement)
			s.Equal("utf8mb4_0900_ai_ci", column.Collation)
			s.Equal("This is a enum column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("enum('1','2','3')", column.Type)
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
			s.Equal("bigint unsigned", column.Type)
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
			s.Equal("int unsigned", column.Type)
			s.Equal("int", column.TypeName)
		}
		if column.Name == "unsigned_big_integer" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Equal("This is a unsigned_big_integer column", column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("bigint unsigned", column.Type)
			s.Equal("bigint", column.TypeName)
		}
	}
}

func (s *SchemaSuite) TestColumnTypes_Sqlserver() {
	if s.driverToTestQuery[sqlserver.Name] == nil {
		s.T().Skip("Skip test")
	}

	testQuery := s.driverToTestQuery[sqlserver.Name]
	schema := newSchema(testQuery, s.driverToTestQuery)
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
		if column.Name == "boolean_default" {
			s.False(column.Autoincrement)
			s.Empty(column.Collation)
			s.Empty(column.Comment)
			s.Equal("('1')", column.Default)
			s.False(column.Nullable)
			s.Equal("bit", column.Type)
			s.Equal("bit", column.TypeName)
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
		if column.Name == "custom_type" {
			s.False(column.Autoincrement)
			s.Empty(column.Comment)
			s.Empty(column.Default)
			s.False(column.Nullable)
			s.Equal("geometry", column.Type)
			s.Equal("geometry", column.TypeName)
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
		if column.Name == "enum_int" {
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

func (s *SchemaSuite) TestEnum_Postgres() {
	if s.driverToTestQuery[postgres.Name] == nil {
		s.T().Skip("Skip test")
	}

	testQuery := s.driverToTestQuery[postgres.Name]
	schema := newSchema(testQuery, s.driverToTestQuery)
	table := "postgres_enum"

	s.NoError(schema.Create(table, func(table contractsschema.Blueprint) {
		table.ID()
		table.Enum("str", []any{"a", "b", "c"})
		table.Enum("int", []any{1, 2, 3})
	}))

	type PostgresEnum struct {
		ID  uint `gorm:"primaryKey"`
		Str string
		Int string
	}

	postgresEnum := &PostgresEnum{
		Str: "a",
		Int: "4",
	}
	s.ErrorContains(testQuery.Query().Table(table).Create(&postgresEnum), `new row for relation "goravel_postgres_enum" violates check constraint "goravel_postgres_enum_int_check"`)

	postgresEnum = &PostgresEnum{
		Str: "a",
		Int: "1",
	}
	s.NoError(testQuery.Query().Table(table).Create(&postgresEnum))
	s.True(postgresEnum.ID > 0)
}

func (s *SchemaSuite) TestEnum_Sqlite() {
	if s.driverToTestQuery[sqlite.Name] == nil {
		s.T().Skip("Skip test")
	}

	testQuery := s.driverToTestQuery[sqlite.Name]
	schema := newSchema(testQuery, s.driverToTestQuery)
	table := "sqlite_enum"

	s.NoError(schema.Create(table, func(table contractsschema.Blueprint) {
		table.ID()
		table.Enum("str", []any{"a", "b", "c"})
		table.Enum("int", []any{1, 2, 3})
	}))

	type SqliteEnum struct {
		ID  uint `gorm:"primaryKey"`
		Str string
		Int string
	}

	sqliteEnum := &SqliteEnum{
		Str: "a",
		Int: "4",
	}
	s.ErrorContains(testQuery.Query().Table(table).Create(&sqliteEnum), `constraint failed: CHECK constraint failed: int`)

	sqliteEnum = &SqliteEnum{
		Str: "a",
		Int: "1",
	}
	s.NoError(testQuery.Query().Table(table).Create(&sqliteEnum))
	s.True(sqliteEnum.ID > 0)
}

func (s *SchemaSuite) TestEnum_Mysql() {
	if s.driverToTestQuery[mysql.Name] == nil {
		s.T().Skip("Skip test")
	}

	testQuery := s.driverToTestQuery[mysql.Name]
	schema := newSchema(testQuery, s.driverToTestQuery)
	table := "mysql_enum"

	s.NoError(schema.Create(table, func(table contractsschema.Blueprint) {
		table.ID()
		table.Enum("str", []any{"a", "b", "c"})
		table.Enum("int", []any{1, 2, 3})
	}))

	type MysqlEnum struct {
		ID  uint `gorm:"primaryKey"`
		Str string
		Int int
	}

	mysqlEnum := &MysqlEnum{
		Str: "a",
		Int: 4,
	}
	s.ErrorContains(testQuery.Query().Table(table).Create(&mysqlEnum), "Data truncated for column 'int' at row 1")

	mysqlEnum = &MysqlEnum{
		Str: "a",
		Int: 1,
	}
	s.NoError(testQuery.Query().Table(table).Create(&mysqlEnum))
	s.True(mysqlEnum.ID > 0)
}

func (s *SchemaSuite) TestEnum_Sqlserver() {
	if s.driverToTestQuery[sqlserver.Name] == nil {
		s.T().Skip("Skip test")
	}

	testQuery := s.driverToTestQuery[sqlserver.Name]
	schema := newSchema(testQuery, s.driverToTestQuery)
	table := "sqlserver_enum"

	s.NoError(schema.Create(table, func(table contractsschema.Blueprint) {
		table.ID()
		table.Enum("str", []any{"a", "b", "c"})
		table.Enum("int", []any{1, 2, 3})
	}))

	type SqlserverEnum struct {
		ID  uint `gorm:"primaryKey"`
		Str string
		Int string
	}

	sqlserverEnum := &SqlserverEnum{
		Str: "a",
		Int: "4",
	}
	s.ErrorContains(testQuery.Query().Table(table).Create(&sqlserverEnum), `The INSERT statement conflicted with the CHECK constraint`)

	sqlserverEnum = &SqlserverEnum{
		Str: "a",
		Int: "1",
	}
	s.NoError(testQuery.Query().Table(table).Create(&sqlserverEnum))
	s.True(sqlserverEnum.ID > 0)
}

func (s *SchemaSuite) TestForeign() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver, func() {
			schema := newSchema(testQuery, s.driverToTestQuery)
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
			})

			s.Require().Nil(err)
			s.Require().True(schema.HasTable(table2))

			table3 := "foreign3"
			err = schema.Create(table3, func(table contractsschema.Blueprint) {
				table.ID()
				table.String("name")
				table.UnsignedBigInteger("foreign1_id")
				table.UnsignedBigInteger("foreign2_id")
				table.Foreign("foreign1_id").References("id").On(table1)
				table.Foreign("foreign2_id").References("id").On(table2).CascadeOnDelete().CascadeOnUpdate().Name("foreign3_foreign2_id_foreign")
			})

			s.Require().Nil(err)
			s.Require().True(schema.HasTable(table3))

			foreignKeys, err := schema.GetForeignKeys(table3)
			s.NoError(err)
			s.Len(foreignKeys, 2)

			for _, foreignKey := range foreignKeys {
				if s.prefix+table1 == foreignKey.ForeignTable {
					s.ElementsMatch([]string{"foreign1_id"}, foreignKey.Columns)
					s.ElementsMatch([]string{"id"}, foreignKey.ForeignColumns)
					s.Equal("no action", foreignKey.OnDelete)
					s.Equal("no action", foreignKey.OnUpdate)
					if driver == sqlite.Name {
						s.Empty(foreignKey.Name)
						s.Empty(foreignKey.ForeignSchema)
					} else {
						s.Equal("goravel_foreign3_foreign1_id_foreign", foreignKey.Name)
						s.NotEmpty(foreignKey.ForeignSchema)
					}
				} else if s.prefix+table2 == foreignKey.ForeignTable {
					s.ElementsMatch([]string{"foreign2_id"}, foreignKey.Columns)
					s.ElementsMatch([]string{"id"}, foreignKey.ForeignColumns)
					s.Equal("cascade", foreignKey.OnDelete)
					s.Equal("cascade", foreignKey.OnUpdate)
					if driver == sqlite.Name {
						s.Empty(foreignKey.Name)
						s.Empty(foreignKey.ForeignSchema)
					} else {
						s.Equal("foreign3_foreign2_id_foreign", foreignKey.Name)
						s.NotEmpty(foreignKey.ForeignSchema)
					}
				} else {
					s.Fail("Unexpected foreign key")
				}
			}

			err = schema.Table(table3, func(table contractsschema.Blueprint) {
				table.DropForeign("foreign1_id")
				table.DropForeignByName("foreign3_foreign2_id_foreign")
			})

			s.NoError(err)

			foreignKeys, err = schema.GetForeignKeys(table3)
			s.NoError(err)
			if driver == sqlite.Name {
				s.Len(foreignKeys, 2)
			} else {
				s.Len(foreignKeys, 0)
			}
		})
	}
}

func (s *SchemaSuite) TestFullText() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver, func() {
			schema := newSchema(testQuery, s.driverToTestQuery)
			table := "fulltext"
			err := schema.Create(table, func(table contractsschema.Blueprint) {
				table.String("name")
				table.String("avatar")
				table.FullText("name")
				table.FullText("avatar").Name("fulltext_avatar_fulltext")
			})

			s.Require().Nil(err)

			if driver == mysql.Name || driver == postgres.Name {
				s.True(schema.HasIndex(table, "goravel_fulltext_name_fulltext"))
				s.True(schema.HasIndex(table, "fulltext_avatar_fulltext"))
			} else {
				s.False(schema.HasIndex(table, "goravel_fulltext_name_fulltext"))
				s.False(schema.HasIndex(table, "fulltext_avatar_fulltext"))
			}

			err = schema.Table(table, func(table contractsschema.Blueprint) {
				table.DropFullText("name")
				table.DropFullTextByName("fulltext_avatar_fulltext")
			})

			s.Require().Nil(err)
			s.False(schema.HasIndex(table, "goravel_fulltext_name_fulltext"))
			s.False(schema.HasIndex(table, "fulltext_avatar_fulltext"))
		})
	}
}

func (s *SchemaSuite) TestGeneratedAs_Postgres() {
	if s.driverToTestQuery[postgres.Name] == nil {
		s.T().Skip("Skip test")
	}

	testQuery := s.driverToTestQuery[postgres.Name]
	schema := newSchema(testQuery, s.driverToTestQuery)
	table := "postgres_generated_as"

	s.NoError(schema.Create(table, func(table contractsschema.Blueprint) {
		table.ID()
		table.String("name")
		table.Integer("small_integer").GeneratedAs()
		table.BigInteger("integer").GeneratedAs("START WITH 10 INCREMENT BY 2")
		table.SmallInteger("big_integer").GeneratedAs("START WITH 20 INCREMENT BY 5").Always()

	}))

	type PostgresGeneratedAs struct {
		ID           uint `gorm:"primaryKey"`
		Name         string
		SmallInteger int
		Integer      int32
		BigInteger   int64
	}

	s.NoError(testQuery.Query().Table(table).Create([]map[string]any{{"name": "test_1"}, {"name": "test_2"}, {"name": "test_3"}}))

	var postgresGeneratedAsList []PostgresGeneratedAs
	s.NoError(testQuery.Query().Table(table).Find(&postgresGeneratedAsList))
	s.Len(postgresGeneratedAsList, 3)

	s.Equal(PostgresGeneratedAs{1, "test_1", 1, 10, 20}, postgresGeneratedAsList[0])
	s.Equal(PostgresGeneratedAs{2, "test_2", 2, 12, 25}, postgresGeneratedAsList[1])
	s.Equal(PostgresGeneratedAs{3, "test_3", 3, 14, 30}, postgresGeneratedAsList[2])
}

func (s *SchemaSuite) TestPrimary() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver, func() {
			schema := newSchema(testQuery, s.driverToTestQuery)
			table := "primaries"

			s.NoError(schema.Create(table, func(table contractsschema.Blueprint) {
				table.String("name")
				table.String("age")
				table.Primary("name", "age")
			}))

			// SQLite does not support set primary index separately
			if driver == postgres.Name {
				s.Require().True(schema.HasIndex(table, "goravel_primaries_pkey"))
			}
			if driver == mysql.Name {
				s.Require().True(schema.HasIndex(table, "primary"))
			}
			if driver == sqlserver.Name {
				s.Require().True(schema.HasIndex(table, "goravel_primaries_name_age_primary"))
			}

			s.NoError(schema.Table(table, func(table contractsschema.Blueprint) {
				table.DropPrimary("name", "age")
			}))
			if driver == postgres.Name {
				s.Require().False(schema.HasIndex(table, "goravel_primaries_pkey"))
			}
			if driver == mysql.Name {
				s.Require().False(schema.HasIndex(table, "primary"))
			}
			if driver == sqlserver.Name {
				s.Require().False(schema.HasIndex(table, "goravel_primaries_name_age_primary"))
			}
		})
	}
}

func (s *SchemaSuite) TestRenameColumn() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver, func() {
			schema := newSchema(testQuery, s.driverToTestQuery)
			table := "rename_column"

			s.NoError(schema.Create(table, func(table contractsschema.Blueprint) {
				table.String("before")
			}))
			s.True(schema.HasColumn(table, "before"))

			s.NoError(schema.Table(table, func(table contractsschema.Blueprint) {
				table.RenameColumn("before", "after")
			}))
			s.False(schema.HasColumn(table, "before"))
			s.True(schema.HasColumn(table, "after"))
		})
	}
}

func (s *SchemaSuite) TestTableComment() {
	for driver, testQuery := range s.driverToTestQuery {
		if driver == sqlite.Name || driver == sqlserver.Name {
			continue
		}
		s.Run(driver, func() {
			schema := newSchema(testQuery, s.driverToTestQuery)
			table := "table_with_comment"
			comment := "It's a table with comment"

			s.NoError(schema.Create(table, func(table contractsschema.Blueprint) {
				table.ID()
				table.Comment(comment)
			}))
			s.True(schema.HasTable(table))

			tables, err := schema.GetTables()
			s.NoError(err)
			for _, t := range tables {
				if t.Name == table {
					s.Equal(comment, t.Comment)
				}
			}
		})
	}
}

func (s *SchemaSuite) TestID_Postgres() {
	if s.driverToTestQuery[postgres.Name] == nil {
		s.T().Skip("Skip test")
	}

	testQuery := s.driverToTestQuery[postgres.Name]
	schema := newSchema(testQuery, s.driverToTestQuery)

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
	if s.driverToTestQuery[sqlite.Name] == nil {
		s.T().Skip("Skip test")
	}

	testQuery := s.driverToTestQuery[sqlite.Name]
	schema := newSchema(testQuery, s.driverToTestQuery)

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
	if s.driverToTestQuery[mysql.Name] == nil {
		s.T().Skip("Skip test")
	}

	testQuery := s.driverToTestQuery[mysql.Name]
	schema := newSchema(testQuery, s.driverToTestQuery)

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
			expectType:     "bigint unsigned",
			expectTypeName: "bigint",
		},
		{
			table: "MediumIncrements",
			setup: func(table string) error {
				return schema.Create(table, func(table contractsschema.Blueprint) {
					table.MediumIncrements("id").Comment("This is a id column")
				})
			},
			expectType:     "mediumint unsigned",
			expectTypeName: "mediumint",
		},
		{
			table: "IntegerIncrements",
			setup: func(table string) error {
				return schema.Create(table, func(table contractsschema.Blueprint) {
					table.IntegerIncrements("id").Comment("This is a id column")
				})
			},
			expectType:     "int unsigned",
			expectTypeName: "int",
		},
		{
			table: "SmallIncrements",
			setup: func(table string) error {
				return schema.Create(table, func(table contractsschema.Blueprint) {
					table.SmallIncrements("id").Comment("This is a id column")
				})
			},
			expectType:     "smallint unsigned",
			expectTypeName: "smallint",
		},
		{
			table: "TinyIncrements",
			setup: func(table string) error {
				return schema.Create(table, func(table contractsschema.Blueprint) {
					table.TinyIncrements("id").Comment("This is a id column")
				})
			},
			expectType:     "tinyint unsigned",
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
	if s.driverToTestQuery[sqlserver.Name] == nil {
		s.T().Skip("Skip test")
	}

	testQuery := s.driverToTestQuery[sqlserver.Name]
	schema := newSchema(testQuery, s.driverToTestQuery)

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
		s.Run(driver, func() {
			schema := newSchema(testQuery, s.driverToTestQuery)
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
					if driver == sqlite.Name {
						s.Empty(index.Type)
					} else if driver == sqlserver.Name {
						s.Equal("nonclustered", index.Type)
					} else {
						s.Equal("btree", index.Type)
					}
					s.False(index.Unique)
				}
				if index.Name == "name_index" {
					s.ElementsMatch(index.Columns, []string{"name"})
					s.False(index.Primary)
					if driver == sqlite.Name {
						s.Empty(index.Type)
					} else if driver == sqlserver.Name {
						s.Equal("nonclustered", index.Type)
					} else {
						s.Equal("btree", index.Type)
					}
					s.False(index.Unique)
				}
				if strings.HasPrefix(index.Name, "pk_") {
					s.ElementsMatch(index.Columns, []string{"id"})
					s.True(index.Primary)
					s.Equal("clustered", index.Type)
					s.True(index.Unique)
				}
				if index.Name == "primary" {
					s.ElementsMatch(index.Columns, []string{"id"})
					s.True(index.Primary)
					if driver == sqlite.Name {
						s.Empty(index.Type)
					} else {
						s.Equal("btree", index.Type)
					}
					s.True(index.Unique)
				}
				if index.Name == "goravel_indexes_pkey" {
					s.ElementsMatch(index.Columns, []string{"id"})
					s.True(index.Primary)
					s.Equal("btree", index.Type)
					s.True(index.Unique)
				}
			}

			s.NoError(schema.Table(table, func(table contractsschema.Blueprint) {
				table.DropIndex("id", "name")
				table.RenameIndex("name_index", "name")
			}))
			s.False(schema.HasIndex(table, "goravel_indexes_id_name_index"))
			s.True(schema.HasIndex(table, "name"))
			s.NoError(schema.Table(table, func(table contractsschema.Blueprint) {
				table.DropIndexByName("name")
			}))
			s.False(schema.HasIndex(table, "name"))
		})
	}
}

func (s *SchemaSuite) TestSql() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver, func() {
			schema := newSchema(testQuery, s.driverToTestQuery)

			s.NoError(schema.Create("sql", func(table contractsschema.Blueprint) {
				table.String("name")
			}))
			s.NoError(schema.Sql("insert into goravel_sql (name) values ('goravel');"))

			count, err := testQuery.Query().Table("sql").Where("name", "goravel").Count()

			s.NoError(err)
			s.Equal(int64(1), count)
		})
	}
}

func (s *SchemaSuite) TestTypeMethods() {
	for driver, testQuery := range s.driverToTestQuery {
		if driver != postgres.Name {
			continue
		}

		s.Run(driver, func() {
			schema := newSchema(testQuery, s.driverToTestQuery)

			s.NoError(schema.Sql("CREATE TYPE person AS (name TEXT, age INT);"))

			s.True(schema.HasType("person"))

			types, err := schema.GetTypes()
			s.NoError(err)
			s.Len(types, 2)

			for _, t := range types {
				if t.Name == "person" {
					s.Equal("person", t.Name)
					s.Equal("composite", t.Type)
					s.Equal("public", t.Schema)
					s.False(t.Implicit)
				}
			}

			s.NoError(schema.DropAllTypes())
			s.False(schema.HasType("person"))
		})
	}
}

func (s *SchemaSuite) TestTableMethods() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver, func() {
			schema := newSchema(testQuery, s.driverToTestQuery)
			tableOne := "table_one"
			tableTwo := "table_two"
			tableThree := "table_three"
			tableFour := "table_four"
			tableFive := "table_five"

			s.Error(schema.Drop(tableOne))
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
			s.NoError(schema.Create(tableFour, func(table contractsschema.Blueprint) {
				table.String("name")
			}))
			s.True(schema.HasTable(tableOne))
			s.True(schema.HasTable(tableTwo))
			s.True(schema.HasTable(tableThree))
			s.True(schema.HasTable(tableFour))

			tables, err := schema.GetTables()

			s.NoError(err)
			s.Len(tables, 4)
			s.ElementsMatch([]string{
				s.prefix + tableOne, s.prefix + tableTwo, s.prefix + tableThree, s.prefix + tableFour,
			}, schema.GetTableListing())

			s.NoError(schema.Rename(tableOne, tableFive))
			s.False(schema.HasTable(tableOne))
			s.True(schema.HasTable(tableFive))

			s.NoError(schema.DropIfExists(tableOne))
			s.False(schema.HasTable(tableOne))

			s.NoError(schema.Table(tableTwo, func(table contractsschema.Blueprint) {
				table.Drop()
			}))
			s.False(schema.HasTable(tableTwo))

			s.NoError(schema.Drop(tableThree))
			s.False(schema.HasTable(tableThree))

			s.NoError(schema.DropAllTables())
			s.False(schema.HasTable(tableFour))
		})
	}
}

func (s *SchemaSuite) TestTimestampMethods() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver, func() {
			schema := newSchema(testQuery, s.driverToTestQuery)
			table := "timestamp"

			s.NoError(schema.Create(table, func(table contractsschema.Blueprint) {
				table.String("name")
				table.Timestamps()
				table.SoftDeletes()
			}))
			s.True(schema.HasColumns(table, []string{"created_at", "updated_at", "deleted_at"}))
			s.NoError(schema.Table(table, func(table contractsschema.Blueprint) {
				table.DropTimestamps()
				table.DropSoftDeletes()
			}))
			s.False(schema.HasColumns(table, []string{"created_at", "updated_at", "deleted_at"}))

			s.NoError(schema.Table(table, func(table contractsschema.Blueprint) {
				table.TimestampsTz()
				table.SoftDeletesTz("delete_at")
			}))
			s.True(schema.HasColumns(table, []string{"created_at", "updated_at", "delete_at"}))
			s.NoError(schema.Table(table, func(table contractsschema.Blueprint) {
				table.DropTimestampsTz()
				table.DropSoftDeletesTz("delete_at")
			}))
			s.False(schema.HasColumns(table, []string{"created_at", "updated_at", "delete_at"}))
		})
	}
}

func (s *SchemaSuite) TestUnique() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver, func() {
			schema := newSchema(testQuery, s.driverToTestQuery)
			table := "uniques"

			s.NoError(schema.Create(table, func(table contractsschema.Blueprint) {
				table.String("name")
				table.String("age")
				table.Unique("name", "age")
			}))

			s.True(schema.HasIndex(table, "goravel_uniques_name_age_unique"))
			s.NoError(schema.Table(table, func(table contractsschema.Blueprint) {
				table.DropUnique("name", "age")
			}))
			s.False(schema.HasIndex(table, "goravel_uniques_name_age_unique"))

			s.NoError(schema.Table(table, func(table contractsschema.Blueprint) {
				table.Unique("name", "age").Name("name_age_unique")
			}))

			s.True(schema.HasIndex(table, "name_age_unique"))
			s.NoError(schema.Table(table, func(table contractsschema.Blueprint) {
				table.DropUniqueByName("name_age_unique")
			}))
			s.False(schema.HasIndex(table, "name_age_unique"))
		})
	}
}

func (s *SchemaSuite) TestViewMethods() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver, func() {
			schema := newSchema(testQuery, s.driverToTestQuery)
			table := "views"

			s.NoError(schema.Create(table, func(table contractsschema.Blueprint) {
				table.String("name")
			}))
			s.NoError(schema.Sql("create view goravel_view as select * from goravel_views;"))
			s.True(schema.HasView("goravel_view"))

			views, err := schema.GetViews()
			s.NoError(err)
			s.Len(views, 1)
			s.Equal("goravel_view", views[0].Name)
			s.NotEmpty(views[0].Definition)

			if driver == postgres.Name || driver == sqlserver.Name {
				s.NotEmpty(views[0].Schema)
			} else {
				s.Empty(views[0].Schema)
			}

			s.NoError(schema.DropAllViews())
			s.False(schema.HasView("goravel_view"))
		})
	}
}

func (s *SchemaSuite) createTableAndAssertColumnsForColumnMethods(schema contractsschema.Schema, table string) {
	err := schema.Create(table, func(table contractsschema.Blueprint) {
		table.BigInteger("big_integer").Comment("This is a big_integer column")
		table.Boolean("boolean_default").Default(true).Comment("This is a boolean column with default value")
		table.Char("char").Comment("This is a char column")
		if schema.GetConnection() != postgres.Name {
			table.Column("custom_type", "geometry").Comment("This is a custom type column")
		} else {
			table.Column("custom_type", "macaddr").Comment("This is a custom type column")
		}
		table.Date("date").Comment("This is a date column")
		table.DateTime("date_time", 3).Comment("This is a date time column")
		table.DateTimeTz("date_time_tz", 3).Comment("This is a date time with time zone column")
		table.Decimal("decimal").Places(1).Total(4).Comment("This is a decimal column")
		table.Double("double").Comment("This is a double column")
		table.Enum("enum", []any{"a", "b", "c"}).Comment("This is a enum column")
		table.Enum("enum_int", []any{1, 2, 3}).Comment("This is a enum column")
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

	s.Equal(35, len(columnListing))
	s.Contains(columnListing, "another_deleted_at")
	s.Contains(columnListing, "big_integer")
	s.Contains(columnListing, "boolean_default")
	s.Contains(columnListing, "char")
	s.Contains(columnListing, "created_at")
	s.Contains(columnListing, "custom_type")
	s.Contains(columnListing, "date")
	s.Contains(columnListing, "date_time")
	s.Contains(columnListing, "date_time_tz")
	s.Contains(columnListing, "decimal")
	s.Contains(columnListing, "deleted_at")
	s.Contains(columnListing, "double")
	s.Contains(columnListing, "enum")
	s.Contains(columnListing, "enum_int")
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

func TestPostgresSchema(t *testing.T) {
	table := "table"
	postgresTestQuery := NewTestQueryBuilder().Postgres("", false)
	postgresTestQuery.WithSchema(testSchema)
	newSchema := newSchema(postgresTestQuery, map[string]*TestQuery{
		postgresTestQuery.Driver().Pool().Writers[0].Connection: postgresTestQuery,
	})

	assert.NoError(t, newSchema.Create(table, func(table contractsschema.Blueprint) {
		table.String("name")
	}))
	tables, err := newSchema.GetTables()

	assert.NoError(t, err)
	assert.Len(t, tables, 1)
	assert.Equal(t, table, tables[0].Name)
	assert.Equal(t, "public", tables[0].Schema)
	assert.True(t, newSchema.HasTable(fmt.Sprintf("public.%s", table)))
	assert.True(t, newSchema.HasTable(table))
}

func TestSqlserverSchema(t *testing.T) {
	schema := "goravel"
	table := "table"
	sqlserverTestQuery := NewTestQueryBuilder().Sqlserver("", false)
	sqlserverTestQuery.WithSchema(testSchema)
	newSchema := newSchema(sqlserverTestQuery, map[string]*TestQuery{
		sqlserverTestQuery.Driver().Pool().Writers[0].Connection: sqlserverTestQuery,
	})

	assert.NoError(t, newSchema.Create(fmt.Sprintf("%s.%s", schema, table), func(table contractsschema.Blueprint) {
		table.String("name")
	}))
	tables, err := newSchema.GetTables()

	assert.NoError(t, err)
	assert.Len(t, tables, 1)
	assert.Equal(t, table, tables[0].Name)
	assert.Equal(t, schema, tables[0].Schema)
	assert.True(t, newSchema.HasTable(fmt.Sprintf("%s.%s", schema, table)))
}

func (s *SchemaSuite) TestExtend() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver, func() {
			schema := newSchema(testQuery, s.driverToTestQuery)

			originalGoTypes := schema.GoTypes()

			customGoTypes := []contractsschema.GoType{
				{Pattern: "uuid", Type: "uuid.UUID", NullType: "uuid.NullUUID", Import: "github.com/google/uuid"},
				{Pattern: "point", Type: "geom.Point", NullType: "*geom.Point", Import: "github.com/twpayne/go-geom"},
				// Override an existing type
				{Pattern: "(?i)^jsonb$", Type: "jsonb.RawMessage", NullType: "*jsonb.RawMessage", Import: "github.com/jmoiron/sqlx/types"},
			}

			schema.Extend(contractsschema.Extension{
				GoTypes: customGoTypes,
			})
			extendedGoTypes := schema.GoTypes()

			s.Equal(len(originalGoTypes)+2, len(extendedGoTypes), "Extended GoTypes list should be longer than original")

			uuidType, found := findGoType("uuid", extendedGoTypes)
			s.True(found, "uuid type should be added")
			s.Equal("uuid.UUID", uuidType.Type)
			s.Equal("uuid.NullUUID", uuidType.NullType)
			s.Equal("github.com/google/uuid", uuidType.Import)

			pointType, found := findGoType("point", extendedGoTypes)
			s.True(found, "point type should be added")
			s.Equal("geom.Point", pointType.Type)
			s.Equal("*geom.Point", pointType.NullType)
			s.Equal("github.com/twpayne/go-geom", pointType.Import)

			// Check that existing type was overridden
			jsonbPattern := "(?i)^jsonb$"
			originalJsonb, found := findGoType(jsonbPattern, originalGoTypes)
			s.True(found, "jsonb type should exist in original types")
			s.Equal("string", originalJsonb.Type, "Original jsonb type should be string")

			extendedJsonb, found := findGoType(jsonbPattern, extendedGoTypes)
			s.True(found, "jsonb type should exist in extended types")
			s.Equal("jsonb.RawMessage", extendedJsonb.Type, "Extended jsonb type should be jsonb.RawMessage")
			s.Equal("*jsonb.RawMessage", extendedJsonb.NullType)
			s.Equal("github.com/jmoiron/sqlx/types", extendedJsonb.Import)
		})
	}
}

// TestEmptyExtend tests that Extend works correctly with empty extensions
func (s *SchemaSuite) TestEmptyExtend() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver, func() {
			schema := newSchema(testQuery, s.driverToTestQuery)

			originalGoTypes := schema.GoTypes()
			originalLength := len(originalGoTypes)

			schema.Extend(contractsschema.Extension{
				GoTypes: []contractsschema.GoType{},
			})

			extendedGoTypes := schema.GoTypes()
			s.Equal(originalLength, len(extendedGoTypes), "GoTypes list should be unchanged after empty extension")

			for i, originalType := range originalGoTypes {
				s.Equal(originalType.Pattern, extendedGoTypes[i].Pattern,
					"Pattern of type at index %d should be unchanged", i)
				s.Equal(originalType.Type, extendedGoTypes[i].Type,
					"Type of type at index %d should be unchanged", i)
				s.Equal(originalType.NullType, extendedGoTypes[i].NullType,
					"NullType of type at index %d should be unchanged", i)
				s.Equal(originalType.Import, extendedGoTypes[i].Import,
					"Imports of type at index %d should be unchanged", i)
				s.Equal(originalType.NullImport, extendedGoTypes[i].NullImport,
					"Import of NullType at index %d should be unchanged", i)
			}
		})
	}
}

func findGoType(pattern string, types []contractsschema.GoType) (contractsschema.GoType, bool) {
	for _, t := range types {
		if t.Pattern == pattern {
			return t, true
		}
	}
	return contractsschema.GoType{}, false
}
