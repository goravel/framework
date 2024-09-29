package console

import (
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/support/file"
)

func createMigrations(driver contractsorm.Driver) {
	switch driver {
	case contractsorm.DriverPostgres:
		createPostgresMigrations()
	case contractsorm.DriverMysql:
		createMysqlMigrations()
	case contractsorm.DriverSqlserver:
		createSqlserverMigrations()
	case contractsorm.DriverSqlite:
		createSqliteMigrations()
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
