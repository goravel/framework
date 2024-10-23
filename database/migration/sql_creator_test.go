package migration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/str"
)

type SqlCreatorSuite struct {
	suite.Suite
	sqlCreator *SqlCreator
}

func TestSqlCreatorSuite(t *testing.T) {
	suite.Run(t, &SqlCreatorSuite{})
}

func (s *SqlCreatorSuite) SetupTest() {
	s.sqlCreator = NewSqlCreator("postgres", "utf8mb4")
}

func (s *SqlCreatorSuite) TestGetPath() {
	now := carbon.FromDateTime(2024, 8, 17, 21, 45, 1)
	carbon.SetTestNow(now)

	pwd, _ := os.Getwd()
	path := filepath.Join(pwd, "database", "migrations")

	s.Equal(filepath.Join(path, "20240817214501_create_users_table.up.sql"), s.sqlCreator.GetPath("create_users_table", "up"))
	s.Equal(filepath.Join(path, "20240817214501_create_users_table.down.sql"), s.sqlCreator.GetPath("create_users_table", "down"))

	carbon.UnsetTestNow()
}

func (s *SqlCreatorSuite) TestPopulateStub() {
	tests := []struct {
		name       string
		driver     database.Driver
		table      string
		create     bool
		expectUp   string
		expectDown string
	}{
		{
			name: "table is empty",
		},
		{
			name:   "mysql - create template",
			driver: "mysql",
			table:  "users",
			create: true,
			expectUp: `CREATE TABLE users (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_users_created_at (created_at),
  KEY idx_users_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;`,
			expectDown: `DROP TABLE IF EXISTS users;`,
		},
		{
			name:       "mysql - update template",
			driver:     "mysql",
			table:      "users",
			create:     false,
			expectUp:   `ALTER TABLE users ADD column varchar(255) COMMENT '';`,
			expectDown: `ALTER TABLE users DROP COLUMN column;`,
		},
		{
			name:   "postgres - create template",
			driver: "postgres",
			table:  "users",
			create: true,
			expectUp: `CREATE TABLE users (
  id SERIAL PRIMARY KEY NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL
);`,
			expectDown: `DROP TABLE IF EXISTS users;`,
		},
		{
			name:       "postgres - update template",
			driver:     "postgres",
			table:      "users",
			create:     false,
			expectUp:   `ALTER TABLE users ADD column varchar(255) NOT NULL;`,
			expectDown: `ALTER TABLE users DROP COLUMN column;`,
		},
		{
			name:   "sqlite - create template",
			driver: "sqlite",
			table:  "users",
			create: true,
			expectUp: `CREATE TABLE users (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL
);`,
			expectDown: `DROP TABLE IF EXISTS users;`,
		},
		{
			name:       "sqlite - update template",
			driver:     "sqlite",
			table:      "users",
			create:     false,
			expectUp:   `ALTER TABLE users ADD column text;`,
			expectDown: `ALTER TABLE users DROP COLUMN column;`,
		},
		{
			name:   "sqlserver - create template",
			driver: "sqlserver",
			table:  "users",
			create: true,
			expectUp: `CREATE TABLE users (
  id bigint NOT NULL IDENTITY(1,1),
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  PRIMARY KEY (id)
);`,
			expectDown: `DROP TABLE IF EXISTS users;`,
		},
		{
			name:       "sqlserver - update template",
			driver:     "sqlserver",
			table:      "users",
			create:     false,
			expectUp:   `ALTER TABLE users ADD column varchar(255);`,
			expectDown: `ALTER TABLE users DROP COLUMN column;`,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			sqlCreator := NewSqlCreator(test.driver, "utf8mb4")
			upStub, downStub := sqlCreator.GetStub(test.table, test.create)
			up := sqlCreator.PopulateStub(upStub, test.table)
			down := sqlCreator.PopulateStub(downStub, test.table)

			s.Equal(test.expectUp, str.Of(up).RTrim("\n").String())
			s.Equal(test.expectDown, str.Of(down).RTrim("\n").String())
		})
	}
}
