package console

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ormcontract "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/database/orm"
	configmock "github.com/goravel/framework/mocks/config"
	consolemock "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/docker"
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/file"
)

type Agent struct {
	orm.Model
	Name string
}

func TestMigrateCommand(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	var (
		mockConfig *configmock.Config
		query      ormcontract.Query
	)

	beforeEach := func() {
		mockConfig = &configmock.Config{}
	}

	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "sqlite",
			setup: func() {
				var err error
				sqliteQuery, err := gorm.NewTestQuery(docker.Sqlite())
				require.NoError(t, err)

				query = sqliteQuery.Query()
				mockConfig = sqliteQuery.MockConfig()
				createSqliteMigrations()
			},
		},
		{
			name: "mysql",
			setup: func() {
				mysqlQuery, err := gorm.NewTestQuery(docker.Mysql())
				require.NoError(t, err)

				query = mysqlQuery.Query()
				mockConfig = mysqlQuery.MockConfig()
				createMysqlMigrations()
			},
		},
		{
			name: "postgres",
			setup: func() {
				var err error
				postgresQuery, err := gorm.NewTestQuery(docker.Postgres())
				assert.Nil(t, err)

				query = postgresQuery.Query()
				mockConfig = postgresQuery.MockConfig()
				createPostgresMigrations()
			},
		},
		{
			name: "sqlserver",
			setup: func() {
				var err error
				sqlserverQuery, err := gorm.NewTestQuery(docker.Sqlserver())
				require.NoError(t, err)

				query = sqlserverQuery.Query()
				mockConfig = sqlserverQuery.MockConfig()
				createSqlserverMigrations()
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			test.setup()

			migrateCommand := NewMigrateCommand(mockConfig)
			mockContext := &consolemock.Context{}
			assert.Nil(t, migrateCommand.Handle(mockContext))

			var agent Agent
			assert.Nil(t, query.Where("name", "goravel").First(&agent))
			assert.True(t, agent.ID > 0)

			removeMigrations()
		})
	}
}

func createMysqlMigrations() {
	_ = file.Create("database/migrations/20230311160527_create_agents_table.up.sql",
		`CREATE TABLE agents (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  name varchar(255) NOT NULL,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_agents_created_at (created_at),
  KEY idx_agents_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
INSERT INTO agents (name, created_at, updated_at) VALUES ('goravel', '2023-03-11 16:07:41', '2023-03-11 16:07:45');
`)
	_ = file.Create("database/migrations/20230311160527_create_agents_table.down.sql",
		`DROP TABLE agents;
`)
}

func createPostgresMigrations() {
	_ = file.Create("database/migrations/20230311160527_create_agents_table.up.sql",
		`CREATE TABLE agents (
  id SERIAL PRIMARY KEY NOT NULL,
  name varchar(255) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL
);
INSERT INTO agents (name, created_at, updated_at) VALUES ('goravel', '2023-03-11 16:07:41', '2023-03-11 16:07:45');
`)
	_ = file.Create("database/migrations/20230311160527_create_agents_table.down.sql",
		`DROP TABLE agents;
`)
}

func createSqlserverMigrations() {
	_ = file.Create("database/migrations/20230311160527_create_agents_table.up.sql",
		`CREATE TABLE agents (
  id bigint NOT NULL IDENTITY(1,1),
  name varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  PRIMARY KEY (id)
);
INSERT INTO agents (name, created_at, updated_at) VALUES ('goravel', '2023-03-11 16:07:41', '2023-03-11 16:07:45');
`)
	_ = file.Create("database/migrations/20230311160527_create_agents_table.down.sql",
		`DROP TABLE agents;
`)
}

func createSqliteMigrations() {
	_ = file.Create("database/migrations/20230311160527_create_agents_table.up.sql",
		`CREATE TABLE agents (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  name varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL
);
INSERT INTO agents (name, created_at, updated_at) VALUES ('goravel', '2023-03-11 16:07:41', '2023-03-11 16:07:45');
`)
	_ = file.Create("database/migrations/20230311160527_create_agents_table.down.sql",
		`DROP TABLE agents;
`)
}

func removeMigrations() {
	_ = file.Remove("database")
	_ = file.Remove("goravel")
}
