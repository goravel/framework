package migration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	mocksconfig "github.com/goravel/framework/mocks/config"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/file"
)

func TestSqlDriverCreate(t *testing.T) {
	var mockConfig *mocksconfig.Config
	now := carbon.FromDateTime(2024, 8, 17, 21, 45, 1)
	carbon.SetTestNow(now)
	pwd, _ := os.Getwd()
	path := filepath.Join(pwd, "database", "migrations")

	beforeEach := func() {
		mockConfig = &mocksconfig.Config{}
	}

	afterEach := func() {
		mockConfig.AssertExpectations(t)
	}

	tests := []struct {
		name        string
		argument    string
		upContent   string
		downContent string
		setup       func()
	}{
		{
			name:        "mysql - empty template",
			argument:    "fix_users_table",
			upContent:   ``,
			downContent: ``,
			setup: func() {
				mockConfig.EXPECT().GetString("database.default").Return("mysql").Twice()
				mockConfig.EXPECT().GetString("database.connections.mysql.charset").Return("utf8mb4").Twice()
			},
		},
		{
			name:     "mysql - create template",
			argument: "create_users_table",
			upContent: `CREATE TABLE users (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_users_created_at (created_at),
  KEY idx_users_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`,
			downContent: `DROP TABLE IF EXISTS users;`,
			setup: func() {
				mockConfig.EXPECT().GetString("database.default").Return("mysql").Times(3)
				mockConfig.EXPECT().GetString("database.connections.mysql.driver").Return("mysql").Once()
				mockConfig.EXPECT().GetString("database.connections.mysql.charset").Return("utf8mb4").Twice()
			},
		},
		{
			name:        "mysql - update template",
			argument:    "add_name_to_users_table",
			upContent:   `ALTER TABLE users ADD column varchar(255) COMMENT ''`,
			downContent: `ALTER TABLE users DROP COLUMN column;`,
			setup: func() {
				mockConfig.EXPECT().GetString("database.default").Return("mysql").Times(3)
				mockConfig.EXPECT().GetString("database.connections.mysql.driver").Return("mysql").Once()
				mockConfig.EXPECT().GetString("database.connections.mysql.charset").Return("utf8mb4").Twice()
			},
		},
		{
			name:        "postgres - empty template",
			argument:    "fix_users_table",
			upContent:   ``,
			downContent: ``,
			setup: func() {
				mockConfig.EXPECT().GetString("database.default").Return("postgres").Twice()
				mockConfig.EXPECT().GetString("database.connections.postgres.charset").Return("utf8mb4").Twice()
			},
		},
		{
			name:     "postgres - create template",
			argument: "create_users_table",
			upContent: `CREATE TABLE users (
  id SERIAL PRIMARY KEY NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL
);
`,
			downContent: `DROP TABLE IF EXISTS users;`,
			setup: func() {
				mockConfig.EXPECT().GetString("database.default").Return("postgres").Times(3)
				mockConfig.EXPECT().GetString("database.connections.postgres.driver").Return("postgres").Once()
				mockConfig.EXPECT().GetString("database.connections.postgres.charset").Return("utf8mb4").Twice()
			},
		},
		{
			name:        "postgres - update template",
			argument:    "add_name_to_users_table",
			upContent:   `ALTER TABLE users ADD column varchar(255) NOT NULL;`,
			downContent: `ALTER TABLE users DROP COLUMN column;`,
			setup: func() {
				mockConfig.EXPECT().GetString("database.default").Return("postgres").Times(3)
				mockConfig.EXPECT().GetString("database.connections.postgres.driver").Return("postgres").Once()
				mockConfig.EXPECT().GetString("database.connections.postgres.charset").Return("utf8mb4").Twice()
			},
		},
		{
			name:        "sqlite - empty template",
			argument:    "fix_users_table",
			upContent:   ``,
			downContent: ``,
			setup: func() {
				mockConfig.EXPECT().GetString("database.default").Return("sqlite").Twice()
				mockConfig.EXPECT().GetString("database.connections.sqlite.charset").Return("utf8mb4").Twice()
			},
		},
		{
			name:     "sqlite - create template",
			argument: "create_users_table",
			upContent: `CREATE TABLE users (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL
);
`,
			downContent: `DROP TABLE IF EXISTS users;`,
			setup: func() {
				mockConfig.EXPECT().GetString("database.default").Return("sqlite").Times(3)
				mockConfig.EXPECT().GetString("database.connections.sqlite.driver").Return("sqlite").Once()
				mockConfig.EXPECT().GetString("database.connections.sqlite.charset").Return("utf8mb4").Twice()
			},
		},
		{
			name:        "sqlite - update template",
			argument:    "add_name_to_users_table",
			upContent:   `ALTER TABLE users ADD column text;`,
			downContent: `ALTER TABLE users DROP COLUMN column;`,
			setup: func() {
				mockConfig.EXPECT().GetString("database.default").Return("sqlite").Times(3)
				mockConfig.EXPECT().GetString("database.connections.sqlite.driver").Return("sqlite").Once()
				mockConfig.EXPECT().GetString("database.connections.sqlite.charset").Return("utf8mb4").Twice()
			},
		},
		{
			name:        "sqlserver - empty template",
			argument:    "fix_users_table",
			upContent:   ``,
			downContent: ``,
			setup: func() {
				mockConfig.EXPECT().GetString("database.default").Return("sqlserver").Twice()
				mockConfig.EXPECT().GetString("database.connections.sqlserver.charset").Return("utf8mb4").Twice()
			},
		},
		{
			name:     "sqlserver - create template",
			argument: "create_users_table",
			upContent: `CREATE TABLE users (
  id bigint NOT NULL IDENTITY(1,1),
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  PRIMARY KEY (id)
);
`,
			downContent: `DROP TABLE IF EXISTS users;`,
			setup: func() {
				mockConfig.EXPECT().GetString("database.default").Return("sqlserver").Times(3)
				mockConfig.EXPECT().GetString("database.connections.sqlserver.driver").Return("sqlserver").Once()
				mockConfig.EXPECT().GetString("database.connections.sqlserver.charset").Return("utf8mb4").Twice()
			},
		},
		{
			name:        "sqlserver - update template",
			argument:    "add_name_to_users_table",
			upContent:   `ALTER TABLE users ADD column varchar(255);`,
			downContent: `ALTER TABLE users DROP COLUMN column;`,
			setup: func() {
				mockConfig.EXPECT().GetString("database.default").Return("sqlserver").Times(3)
				mockConfig.EXPECT().GetString("database.connections.sqlserver.driver").Return("sqlserver").Once()
				mockConfig.EXPECT().GetString("database.connections.sqlserver.charset").Return("utf8mb4").Twice()
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			test.setup()

			driver := &SqlDriver{config: mockConfig}

			assert.Nil(t, driver.Create(test.argument))
			assert.True(t, file.Exists(filepath.Join(path, "20240817214501_"+test.argument+".up.sql")))
			assert.True(t, file.Exists(filepath.Join(path, "20240817214501_"+test.argument+".down.sql")))
			assert.True(t, file.Contain(driver.getPath(test.argument, "up"), test.upContent))
			assert.True(t, file.Contain(driver.getPath(test.argument, "down"), test.downContent))

			afterEach()
		})
	}

	assert.Nil(t, file.Remove("database"))
}
