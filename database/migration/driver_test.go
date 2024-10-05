package migration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/file"
)

type DriverSuite struct {
	suite.Suite
	drivers map[string]migration.Driver
}

func TestDriverSuite(t *testing.T) {
	suite.Run(t, &DriverSuite{})
}

func (s *DriverSuite) SetupTest() {
	s.drivers = map[string]migration.Driver{
		migration.DriverDefault: NewDefaultDriver(),
		migration.DriverSql:     NewSqlDriver("postgres", "utf8mb4"),
	}
}

func (s *DriverSuite) TestCreate() {
	now := carbon.FromDateTime(2024, 8, 17, 21, 45, 1)
	carbon.SetTestNow(now)

	pwd, _ := os.Getwd()
	path := filepath.Join(pwd, "database", "migrations")
	name := "create_users_table"

	for driverName, driver := range s.drivers {
		s.Run(driverName, func() {
			s.NoError(driver.Create(name))

			if driverName == migration.DriverDefault {
				migrationFile := filepath.Join(path, "20240817214501_"+name+".go")
				s.True(file.Exists(migrationFile))
			}

			if driverName == migration.DriverSql {
				upFile := filepath.Join(path, "20240817214501_"+name+".up.sql")
				downFile := filepath.Join(path, "20240817214501_"+name+".down.sql")

				s.True(file.Exists(upFile))
				s.True(file.Exists(downFile))
			}
		})
	}

	defer func() {
		carbon.UnsetTestNow()
		s.NoError(file.Remove("database"))
	}()
}
