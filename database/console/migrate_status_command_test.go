package console

import (
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"

	configmock "github.com/goravel/framework/contracts/config/mocks"
	consolemocks "github.com/goravel/framework/contracts/console/mocks"
	ormcontract "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/gorm"
)

func TestMigrateStatusCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping tests of using docker")
	}

	var (
		mockConfig *configmock.Config
		pool       *dockertest.Pool
		resource   *dockertest.Resource
		query      ormcontract.Query
	)

	beforeEach := func() {
		pool = nil
		mockConfig = &configmock.Config{}
	}

	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "mysql",
			setup: func() {
				var err error
				docker := gorm.NewMysqlDocker()
				pool, resource, query, err = docker.New()
				assert.Nil(t, err)
				mockConfig = docker.MockConfig
				createMysqlMigrations()

			},
		},
		{
			name: "postgresql",
			setup: func() {
				var err error
				docker := gorm.NewPostgresqlDocker()
				pool, resource, query, err = docker.New()
				assert.Nil(t, err)
				mockConfig = docker.MockConfig
				createPostgresqlMigrations()
			},
		},
		{
			name: "sqlserver",
			setup: func() {
				var err error
				docker := gorm.NewSqlserverDocker()
				pool, resource, query, err = docker.New()
				assert.Nil(t, err)
				mockConfig = docker.MockConfig
				createSqlserverMigrations()
			},
		},
		{
			name: "sqlite",
			setup: func() {
				var err error
				docker := gorm.NewSqliteDocker("goravel")
				pool, resource, query, err = docker.New()
				assert.Nil(t, err)
				mockConfig = docker.MockConfig
				createSqliteMigrations()
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			test.setup()

			mockContext := &consolemocks.Context{}

			migrateCommand := NewMigrateCommand(mockConfig)
			assert.Nil(t, migrateCommand.Handle(mockContext))

			migrateStatusCommand := NewMigrateStatusCommand(mockConfig)
			assert.Nil(t, migrateStatusCommand.Handle(mockContext))

			res, err := query.Table("migrations").Where("dirty", false).Update("dirty", true)
			assert.Nil(t, err)
			assert.Equal(t, int64(1), res.RowsAffected)

			assert.Nil(t, migrateStatusCommand.Handle(mockContext))

			if pool != nil && test.name != "sqlite" {
				assert.Nil(t, pool.Purge(resource))
			}

			removeMigrations()
		})
	}
}
