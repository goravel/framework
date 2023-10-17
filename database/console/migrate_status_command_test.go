package console

import (
	"os"
	"testing"

	"github.com/gookit/color"
	"github.com/stretchr/testify/assert"

	ormcontract "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/gorm"
	configmock "github.com/goravel/framework/mocks/config"
	consolemocks "github.com/goravel/framework/mocks/console"
)

func TestMigrateStatusCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping tests of using docker")
	}
	if len(os.Getenv("GORAVEL_DOCKER_TEST")) == 0 {
		color.Redln("Skip tests because not set GORAVEL_DOCKER_TEST environment variable")
		return
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
			name: "mysql",
			setup: func() {
				var err error
				docker := gorm.NewMysqlDocker()
				query, err = docker.New()
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
				query, err = docker.New()
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
				query, err = docker.New()
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
				query, err = docker.New()
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

			removeMigrations()
		})
	}
}
