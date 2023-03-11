package console

import (
	"testing"

	"github.com/stretchr/testify/assert"

	consolemocks "github.com/goravel/framework/contracts/console/mocks"
	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/support/file"
)

type Agent struct {
	orm.Model
	Name string
}

func TestMigrateCommand(t *testing.T) {
	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "mysql",
			setup: func() {
				mysqlPool, mysqlResource, mysqlQuery, err := gorm.MysqlDocker()
				assert.Nil(t, err)

				createMysqlMigrations()
				migrateCommand := &MigrateCommand{}
				mockContext := &consolemocks.Context{}
				err = migrateCommand.Handle(mockContext)
				assert.Nil(t, err)

				var agent Agent
				err = mysqlQuery.Where("name", "goravel").First(&agent)
				assert.Nil(t, err)
				assert.True(t, agent.ID > 0)

				removeMigrations()
				err = mysqlPool.Purge(mysqlResource)
				assert.Nil(t, err)
			},
		},
		{
			name: "postgresql",
			setup: func() {
				postgresqlPool, postgresqlResource, postgresqlDB, err := gorm.PostgresqlDocker()
				assert.Nil(t, err)

				createPostgresqlMigrations()
				migrateCommand := &MigrateCommand{}
				mockContext := &consolemocks.Context{}
				err = migrateCommand.Handle(mockContext)
				assert.Nil(t, err)

				var agent Agent
				err = postgresqlDB.Where("name", "goravel").First(&agent)
				assert.Nil(t, err)
				assert.True(t, agent.ID > 0)

				removeMigrations()
				err = postgresqlPool.Purge(postgresqlResource)
				assert.Nil(t, err)
			},
		},
		{
			name: "sqlserver",
			setup: func() {
				sqlserverPool, sqlserverResource, sqlserverDB, err := gorm.SqlserverDocker()
				assert.Nil(t, err)

				createSqlserverMigrations()
				migrateCommand := &MigrateCommand{}
				mockContext := &consolemocks.Context{}
				err = migrateCommand.Handle(mockContext)
				assert.Nil(t, err)

				var agent Agent
				err = sqlserverDB.Where("name", "goravel").First(&agent)
				assert.Nil(t, err)
				assert.True(t, agent.ID > 0)

				removeMigrations()
				err = sqlserverPool.Purge(sqlserverResource)
				assert.Nil(t, err)
			},
		},
		{
			name: "sqlite",
			setup: func() {
				_, _, sqliteDB, err := gorm.SqliteDocker("goravel")
				assert.Nil(t, err)

				createSqliteMigrations()
				migrateCommand := &MigrateCommand{}
				mockContext := &consolemocks.Context{}
				err = migrateCommand.Handle(mockContext)
				assert.Nil(t, err)

				var agent Agent
				err = sqliteDB.Where("name", "goravel").First(&agent)
				assert.Nil(t, err)
				assert.True(t, agent.ID > 0)

				removeMigrations()
				file.Remove("goravel")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setup()
		})
	}
}

func createMysqlMigrations() {
	file.Create("database/migrations/20230311160527_create_agents_table.up.sql",
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
}

func createPostgresqlMigrations() {
	file.Create("database/migrations/20230311160527_create_agents_table.up.sql",
		`CREATE TABLE agents (
  id SERIAL PRIMARY KEY NOT NULL,
  name varchar(255) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL
);
INSERT INTO agents (name, created_at, updated_at) VALUES ('goravel', '2023-03-11 16:07:41', '2023-03-11 16:07:45');
`)
}

func createSqlserverMigrations() {
	file.Create("database/migrations/20230311160527_create_agents_table.up.sql",
		`CREATE TABLE agents (
  id bigint NOT NULL IDENTITY(1,1),
  name varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  PRIMARY KEY (id)
);
INSERT INTO agents (name, created_at, updated_at) VALUES ('goravel', '2023-03-11 16:07:41', '2023-03-11 16:07:45');
`)
}

func createSqliteMigrations() {
	file.Create("database/migrations/20230311160527_create_agents_table.up.sql",
		`CREATE TABLE agents (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  name varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL
);
INSERT INTO agents (name, created_at, updated_at) VALUES ('goravel', '2023-03-11 16:07:41', '2023-03-11 16:07:45');
`)
}

func removeMigrations() {
	file.Remove("database")
}
