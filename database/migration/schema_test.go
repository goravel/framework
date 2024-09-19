package migration

import (
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database/migration"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	contractstesting "github.com/goravel/framework/contracts/testing"
	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/database/migration/grammars"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksmigration "github.com/goravel/framework/mocks/database/migration"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
	mockslog "github.com/goravel/framework/mocks/log"
	supportdocker "github.com/goravel/framework/support/docker"
	"github.com/goravel/framework/support/env"
)

type TestDB struct {
	config contractstesting.DatabaseConfig
	query  contractsorm.Query
}

type SchemaSuite struct {
	suite.Suite
	driverToTestDB map[contractsorm.Driver]TestDB

	postgresDriver contractstesting.DatabaseDriver
	postgresQuery  contractsorm.Query

	mockConfig *mocksconfig.Config
	mockOrm    *mocksorm.Orm
	mockLog    *mockslog.Log
	schema     *Schema
}

func TestSchemaSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	suite.Run(t, &SchemaSuite{})
}

func (s *SchemaSuite) SetupSuite() {
	postgresDriver := supportdocker.Postgres()
	postgresDocker := gorm.NewPostgresDocker(postgresDriver)
	postgresqlQuery, err := postgresDocker.New()
	if err != nil {
		log.Fatalf("Init postgres docker error: %v", err)
	}

	s.postgresDriver = postgresDriver
	s.postgresQuery = postgresqlQuery
}

func (s *SchemaSuite) SetupTest() {
	mockBlueprint := mocksmigration.NewBlueprint(s.T())
	mockConfig := mocksconfig.NewConfig(s.T())
	mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.driver", "mysql")).
		Return("mysql").Once()
	mockLog := mockslog.NewLog(s.T())
	mockOrm := mocksorm.NewOrm(s.T())

	schema, err := NewSchema(mockBlueprint, mockConfig, "mysql", mockLog, mockOrm)
	s.Nil(err)

	s.mockConfig = mockConfig
	s.mockLog = mockLog
	s.mockOrm = mockOrm
	s.schema = schema

	s.driverToTestDB = map[contractsorm.Driver]TestDB{
		contractsorm.DriverPostgres: {
			config: s.postgresDriver.Config(),
			query:  s.postgresQuery,
		},
	}
}

func (s *SchemaSuite) TestConnection() {
	for driver, _ := range s.driverToTestDB {
		s.mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.schema", driver.String())).
			Return("").Once()
		s.mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.prefix", driver.String())).
			Return("").Once()
		s.mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.driver", driver.String())).
			Return(driver.String()).Once()

		s.NotNil(s.schema.Connection(driver.String()))
	}
}

func (s *SchemaSuite) TestCreate() {
	for driver, testDB := range s.driverToTestDB {
		s.Run(driver.String(), func() {
			s.mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.schema", driver.String())).
				Return("").Once()
			s.mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.prefix", driver.String())).
				Return("").Once()
			s.mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.driver", driver.String())).
				Return(driver.String()).Once()

			mockOrm := mocksorm.NewOrm(s.T())
			s.mockOrm.EXPECT().Connection(driver.String()).Return(mockOrm).Once()
			mockOrm.EXPECT().Query().Return(testDB.query).Once()

			mockColumnDefinition := mocksmigration.NewColumnDefinition(s.T())
			//mockGrammar := mocksmigration.NewGrammar(s.T())

			mockBlueprint := mocksmigration.NewBlueprint(s.T())
			mockBlueprint.EXPECT().SetTable("creates").Once()
			mockBlueprint.EXPECT().Create().Once()
			mockBlueprint.EXPECT().String("name").Return(mockColumnDefinition).Once()
			mockBlueprint.EXPECT().Build(testDB.query, mock.MatchedBy(func(grammar migration.Grammar) bool {
				return s.Equal(reflect.TypeOf(getGrammar(driver)).Elem(), reflect.TypeOf(grammar).Elem())
			})).Return(nil).Once()

			schema := s.schema.Connection(driver.String()).(*Schema)
			schema.blueprint = mockBlueprint
			//schema.grammar = mockGrammar
			schema.Create("creates", func(table migration.Blueprint) {
				table.String("name")

				// TODO Open below when implementing Comment
				//table.Comment("This is a test table")
			})

			// TODO Open below when implementing HasTable
			//s.True(schema.schema.HasTable("creates"))
		})
	}
}

func (s *SchemaSuite) TestDropIfExists() {
	for driver, _ := range s.driverToTestDB {
		s.Run(driver.String(), func() {
			table := "drop_if_exists"

			s.schema.DropIfExists(table)

			s.schema.Create(table, func(table migration.Blueprint) {
				table.String("name")
			})

			// TODO Open below when implementing HasTable
			//s.True(schema.schema.HasTable(table))

			s.schema.DropIfExists(table)

			// TODO Open below when implementing HasTable
			//s.False(schema.schema.HasTable(table))
		})
	}
}

func (s *SchemaSuite) TestInitGrammarAndProcess() {
	for driver, _ := range s.driverToTestDB {
		s.Run(driver.String(), func() {
			s.Nil(s.schema.initGrammar())
			grammarType := reflect.TypeOf(s.schema.grammar)
			grammarName := grammarType.Elem().Name()

			// TODO Open below when implementing Processor
			//processorType := reflect.TypeOf(schema.schema.processor)
			//processorName := processorType.Elem().Name()

			switch driver {
			case contractsorm.DriverMysql:
				s.Equal("Mysql", grammarName)
				//s.Equal("Mysql", processorName)
			case contractsorm.DriverPostgres:
				s.Equal("Postgres", grammarName)
				//s.Equal("Postgres", processorName)
			case contractsorm.DriverSqlserver:
				s.Equal("Sqlserver", grammarName)
				//s.Equal("Sqlserver", processorName)
			case contractsorm.DriverSqlite:
				s.Equal("Sqlite", grammarName)
				//s.Equal("Sqlite", processorName)
			default:
				s.Fail("unsupported database driver")
			}
		})
	}
}

func getGrammar(driver contractsorm.Driver) migration.Grammar {
	switch driver {
	case contractsorm.DriverMysql:
		return nil
	case contractsorm.DriverPostgres:
		return grammars.NewPostgres()
	case contractsorm.DriverSqlserver:
		return nil
	case contractsorm.DriverSqlite:
		return nil
	default:
		return nil
	}
}
