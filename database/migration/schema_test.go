package migration

import (
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database/migration"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/gorm"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
	mockslog "github.com/goravel/framework/mocks/log"
	supportdocker "github.com/goravel/framework/support/docker"
	"github.com/goravel/framework/support/env"
)

type SchemaSuite struct {
	suite.Suite
	driverToTestDB map[contractsorm.Driver]contractsorm.Query
}

func TestSchemaSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	suite.Run(t, &SchemaSuite{})
}

func (s *SchemaSuite) SetupTest() {
	postgresDriver := supportdocker.Postgres()
	postgresDocker := gorm.NewPostgresDocker(postgresDriver)
	postgresQuery, err := postgresDocker.New()
	if err != nil {
		log.Fatalf("Init postgres docker error: %v", err)
	}

	s.driverToTestDB = map[contractsorm.Driver]contractsorm.Query{
		contractsorm.DriverPostgres: postgresQuery,
	}
}

func (s *SchemaSuite) TestDropIfExists() {
	for driver, query := range s.driverToTestDB {
		s.Run(driver.String(), func() {
			schema, _, _, mockOrm := initSchema(s.T(), driver)

			table := "drop_if_exists"

			mockOrm.EXPECT().Connection(schema.connection).Return(mockOrm).Twice()
			mockOrm.EXPECT().Query().Return(query).Twice()
			schema.DropIfExists(table)
			schema.Create(table, func(table migration.Blueprint) {
				table.String("name")
			})

			mockOrm.EXPECT().Query().Return(query).Once()
			s.True(schema.HasTable(table))

			mockOrm.EXPECT().Connection(schema.connection).Return(mockOrm).Once()
			mockOrm.EXPECT().Query().Return(query).Once()
			schema.DropIfExists(table)

			mockOrm.EXPECT().Query().Return(query).Once()
			s.False(schema.HasTable(table))
		})
	}
}

func initSchema(t *testing.T, driver contractsorm.Driver) (*Schema, *mocksconfig.Config, *mockslog.Log, *mocksorm.Orm) {
	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.driver", driver)).Return(driver.String()).Once()
	mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.prefix", driver)).Return("goravel_").Once()
	mockLog := mockslog.NewLog(t)
	mockOrm := mocksorm.NewOrm(t)

	schema := NewSchema(mockConfig, driver.String(), mockLog, mockOrm)

	return schema, mockConfig, mockLog, mockOrm
}
