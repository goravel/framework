package migration

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	contractsdatabase "github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/migration"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	contractstesting "github.com/goravel/framework/contracts/testing"
	"github.com/goravel/framework/database/gorm"
	mocksconfig "github.com/goravel/framework/mocks/config"
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
	driverToTestDB map[contractsdatabase.Driver]TestDB
}

func TestSchemaSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	suite.Run(t, &SchemaSuite{})
}

func (s *SchemaSuite) SetupSuite() {
	postgresDocker := supportdocker.Postgres()
	postgresQuery := gorm.NewTestQuery(postgresDocker)
	s.driverToTestDB = map[contractsdatabase.Driver]TestDB{
		contractsdatabase.DriverPostgres: {
			config: postgresDocker.Config(),
			query:  postgresQuery.Query(),
		},
	}
}

func (s *SchemaSuite) SetupTest() {

}

func (s *SchemaSuite) TestConnection() {
	schema, mockConfig, _, _ := initTest(s.T(), contractsdatabase.DriverMysql)
	connection := contractsdatabase.DriverPostgres.String()
	mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.prefix", connection)).Return("goravel_").Once()
	mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.schema", connection)).Return("").Once()
	mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.driver", connection)).Return(connection).Once()

	s.NotNil(schema.Connection(connection))

	// TODO Test the new schema is valid when implementing HasTable
}

func (s *SchemaSuite) TestDropIfExists() {
	for driver, testDB := range s.driverToTestDB {
		s.Run(driver.String(), func() {
			schema, _, _, mockOrm := initTest(s.T(), driver)

			table := "drop_if_exists"

			mockOrm.EXPECT().Connection(schema.connection).Return(mockOrm).Twice()
			mockOrm.EXPECT().Query().Return(testDB.query).Twice()

			schema.DropIfExists(table)

			schema.Create(table, func(table migration.Blueprint) {
				table.String("name")
			})

			// TODO Open below when implementing HasTable
			//s.True(schema.schema.HasTable(table))
			//s.schema.DropIfExists(table)
			//s.False(schema.schema.HasTable(table))
		})
	}
}

func initTest(t *testing.T, driver contractsdatabase.Driver) (*Schema, *mocksconfig.Config, *mockslog.Log, *mocksorm.Orm) {
	blueprint := NewBlueprint("goravel_", "")
	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.driver", driver)).
		Return(driver.String()).Once()
	mockLog := mockslog.NewLog(t)
	mockOrm := mocksorm.NewOrm(t)

	schema := NewSchema(blueprint, mockConfig, driver.String(), mockLog, mockOrm)

	return schema, mockConfig, mockLog, mockOrm
}
