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
	postgresTestQuery := postgresTestQuery("goravel_", true)
	mysqlTestQuery := mysqlTestQuery("goravel_", true)
	sqlserverTestQuery := sqlserverTestQuery("goravel_", true)
	sqliteTestQuery := sqliteTestQuery("goravel_", true)

	s.driverToTestQuery = map[string]*TestQuery{
		postgresTestQuery.Driver().Config().Driver:  postgresTestQuery,
		mysqlTestQuery.Driver().Config().Driver:     mysqlTestQuery,
		sqlserverTestQuery.Driver().Config().Driver: sqlserverTestQuery,
		sqliteTestQuery.Driver().Config().Driver:    sqliteTestQuery,
	}
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

			s.NoError(migrator.Run())
			s.True(schema.HasTable("users"))

			s.NoError(migrator.Reset())
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

			s.NoError(migrator.Run())
			s.True(schema.HasTable("users"))

			s.NoError(migrator.Rollback(1, 0))
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
	postgresTestQuery := postgresTestQuery("", false)
	postgresTestQuery.WithSchema("goravel")

	schema := newSchema(postgresTestQuery, map[string]*TestQuery{
		postgresTestQuery.Driver().Config().Driver: postgresTestQuery,
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
	sqlserverTestQuery := sqlserverTestQuery("", false)
	sqlserverTestQuery.WithSchema("goravel")

	schema := newSchema(sqlserverTestQuery, map[string]*TestQuery{
		sqlserverTestQuery.Driver().Config().Driver: sqlserverTestQuery,
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
