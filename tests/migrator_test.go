package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	contractsmigration "github.com/goravel/framework/contracts/database/migration"
	contractsschema "github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/database/migration"
	"github.com/goravel/sqlite"
)

type DefaultMigratorWithDBSuite struct {
	suite.Suite
	driverToTestQuery map[string]*TestQuery
}

func TestDefaultMigratorWithDBSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &DefaultMigratorWithDBSuite{})
}

func (s *DefaultMigratorWithDBSuite) SetupTest() {
	s.driverToTestQuery = NewTestQueryBuilder().All("goravel_", true)
}

func (s *DefaultMigratorWithDBSuite) TearDownTest() {
	if s.driverToTestQuery[sqlite.Name] != nil {
		docker, err := s.driverToTestQuery[sqlite.Name].Driver().Docker()
		s.NoError(err)
		s.NoError(docker.Shutdown())
	}
}

func (s *DefaultMigratorWithDBSuite) TestRun() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver, func() {
			schema := newSchema(testQuery, s.driverToTestQuery)
			testMigration := NewTestMigration(schema)
			schema.Register([]contractsschema.Migration{
				testMigration,
			})

			migrator := migration.NewMigrator(nil, schema, "migrations")

			s.NoError(migrator.Run())
			s.True(schema.HasTable("users"))
			status, err := migrator.Status()
			s.NoError(err)
			s.Len(status, 1)
		})
	}
}

func (s *DefaultMigratorWithDBSuite) TestReset() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver, func() {
			schema := newSchema(testQuery, s.driverToTestQuery)
			testMigration := NewTestMigration(schema)
			schema.Register([]contractsschema.Migration{
				testMigration,
			})

			migrator := migration.NewMigrator(nil, schema, "migrations")

			s.NoError(migrator.Reset())
			s.NoError(migrator.Run())
			s.True(schema.HasTable("users"))
			s.NoError(migrator.Reset())
			s.False(schema.HasTable("users"))
		})
	}
}

func (s *DefaultMigratorWithDBSuite) TestRollback() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver, func() {
			schema := newSchema(testQuery, s.driverToTestQuery)
			testMigration := NewTestMigration(schema)
			schema.Register([]contractsschema.Migration{
				testMigration,
			})

			migrator := migration.NewMigrator(nil, schema, "migrations")

			s.NoError(migrator.Rollback(1, 0))
			s.NoError(migrator.Run())
			s.True(schema.HasTable("users"))
			s.NoError(migrator.Rollback(1, 0))
			s.False(schema.HasTable("users"))
		})
	}
}

func (s *DefaultMigratorWithDBSuite) TestStatus() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver, func() {
			schema := newSchema(testQuery, s.driverToTestQuery)
			testMigration := NewTestMigration(schema)
			migrator := migration.NewMigrator(nil, schema, "migrations")
			status, err := migrator.Status()
			s.NoError(err)
			s.Len(status, 0)

			schema.Register([]contractsschema.Migration{
				testMigration,
			})

			s.NoError(migrator.Run())
			s.True(schema.HasTable("users"))
			status, err = migrator.Status()
			s.NoError(err)
			s.Equal(status, []contractsmigration.Status{
				{
					Name:  testMigration.Signature(),
					Batch: 1,
					Ran:   true,
				},
			})
		})
	}
}

func TestDefaultMigratorWithPostgresSchema(t *testing.T) {
	postgresTestQuery := NewTestQueryBuilder().Postgres("", false)
	postgresTestQuery.WithSchema("goravel")
	driverName := postgresTestQuery.Driver().Pool().Writers[0].Driver

	schema := newSchema(postgresTestQuery, map[string]*TestQuery{
		driverName: postgresTestQuery,
	})
	testMigration := NewTestMigration(schema)
	schema.Register([]contractsschema.Migration{
		testMigration,
	})
	migrator := migration.NewMigrator(nil, schema, "migrations")

	assert.NoError(t, migrator.Run())
	assert.True(t, schema.HasTable("users"))
	assert.NoError(t, migrator.Rollback(1, 0))
	assert.False(t, schema.HasTable("users"))
}

func TestDefaultMigratorWithSqlserverSchema(t *testing.T) {
	sqlserverTestQuery := NewTestQueryBuilder().Sqlserver("", false)
	sqlserverTestQuery.WithSchema("goravel")
	driverName := sqlserverTestQuery.Driver().Pool().Writers[0].Driver

	schema := newSchema(sqlserverTestQuery, map[string]*TestQuery{
		driverName: sqlserverTestQuery,
	})
	testMigration := NewTestMigrationWithSqlserverSchema(schema)
	schema.Register([]contractsschema.Migration{
		testMigration,
	})
	migrator := migration.NewMigrator(nil, schema, "migrations")

	assert.NoError(t, migrator.Run())
	assert.True(t, schema.HasTable("goravel.users"))
	assert.NoError(t, migrator.Rollback(1, 0))
	assert.False(t, schema.HasTable("goravel.users"))
}

func TestMigratorWithNonDefaultConnection(t *testing.T) {
	postgresTestQuery := NewTestQueryBuilder().Postgres("", false)
	sqliteTestQuery := NewTestQueryBuilder().Sqlite("", false)

	docker, err := sqliteTestQuery.Driver().Docker()
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, docker.Shutdown())
	}()

	sqliteConfig := sqliteTestQuery.Driver().Pool().Writers[0]
	sqliteConnection := sqliteConfig.Connection
	postgresConnection := postgresTestQuery.Driver().Pool().Writers[0].Connection

	// Register the sqlite connection on the default (postgres) config so it can be
	// lazily resolved when a migration requests it via Connection().
	mockDatabaseConfig(postgresTestQuery.MockConfig(), sqliteConfig)

	// Only the default connection is pre-cached; sqlite is loaded lazily (cold path).
	schema := newSchema(postgresTestQuery, map[string]*TestQuery{
		postgresConnection: postgresTestQuery,
	})

	schema.Register([]contractsschema.Migration{
		NewTestConnectionMigration(schema, sqliteConnection, "20260826160940_create_users_table", "users"),
		NewTestConnectionMigration(schema, sqliteConnection, "20260826161140_create_user_tokens_table", "user_tokens"),
	})

	migrator := migration.NewMigrator(nil, schema, "migrations")

	// Fails on current code: the 2nd ledger row is written to sqlite (no migrations table).
	assert.NoError(t, migrator.Run())

	// Both tables were created on the non-default (sqlite) connection.
	assert.True(t, schema.Connection(sqliteConnection).HasTable("users"))
	assert.True(t, schema.Connection(sqliteConnection).HasTable("user_tokens"))

	// Both ledger rows were written to the default (postgres) connection.
	status, err := migrator.Status()
	assert.NoError(t, err)
	assert.Len(t, status, 2)
	for _, s := range status {
		assert.True(t, s.Ran)
	}
}

type TestMigration struct {
	schema contractsschema.Schema
}

func NewTestMigration(schema contractsschema.Schema) *TestMigration {
	return &TestMigration{schema: schema}
}

func (r *TestMigration) Signature() string {
	return "20240817214501_create_users_table"
}

func (r *TestMigration) Up() error {
	return r.schema.Create("users", func(table contractsschema.Blueprint) {
		table.String("name")
	})
}

func (r *TestMigration) Down() error {
	return r.schema.DropIfExists("users")
}

type TestMigrationWithSqlserverSchema struct {
	schema contractsschema.Schema
}

func NewTestMigrationWithSqlserverSchema(schema contractsschema.Schema) *TestMigrationWithSqlserverSchema {
	return &TestMigrationWithSqlserverSchema{schema: schema}
}

func (r *TestMigrationWithSqlserverSchema) Signature() string {
	return "20240817214501_create_users_table"
}

func (r *TestMigrationWithSqlserverSchema) Up() error {
	return r.schema.Create("goravel.users", func(table contractsschema.Blueprint) {
		table.String("name")
	})
}

func (r *TestMigrationWithSqlserverSchema) Down() error {
	return r.schema.DropIfExists("goravel.users")
}

type TestConnectionMigration struct {
	schema     contractsschema.Schema
	connection string
	signature  string
	table      string
}

func NewTestConnectionMigration(schema contractsschema.Schema, connection, signature, table string) *TestConnectionMigration {
	return &TestConnectionMigration{schema: schema, connection: connection, signature: signature, table: table}
}

func (r *TestConnectionMigration) Signature() string {
	return r.signature
}

func (r *TestConnectionMigration) Connection() string {
	return r.connection
}

func (r *TestConnectionMigration) Up() error {
	return r.schema.Create(r.table, func(table contractsschema.Blueprint) {
		table.String("name")
	})
}

func (r *TestConnectionMigration) Down() error {
	return r.schema.DropIfExists(r.table)
}
