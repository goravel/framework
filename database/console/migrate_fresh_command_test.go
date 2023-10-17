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

func TestMigrateFreshCommand(t *testing.T) {
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
			mockArtisan := &consolemocks.Artisan{}
			migrateCommand := NewMigrateCommand(mockConfig)
			assert.Nil(t, migrateCommand.Handle(mockContext))
			mockContext.On("OptionBool", "seed").Return(false).Once()
			migrateFreshCommand := NewMigrateFreshCommand(mockConfig, mockArtisan)
			assert.Nil(t, migrateFreshCommand.Handle(mockContext))

			var agent Agent
			err := query.Where("name", "goravel").First(&agent)
			assert.Nil(t, err)
			assert.True(t, agent.ID > 0)

			// Test MigrateFreshCommand with --seed flag and seeders specified
			mockContext = &consolemocks.Context{}
			mockArtisan = &consolemocks.Artisan{}
			mockContext.On("OptionBool", "seed").Return(true).Once()
			mockContext.On("OptionSlice", "seeder").Return([]string{"MockSeeder"}).Once()
			mockArtisan.On("Call", "db:seed --seeder MockSeeder").Return(nil).Once()
			migrateFreshCommand = NewMigrateFreshCommand(mockConfig, mockArtisan)
			assert.Nil(t, migrateFreshCommand.Handle(mockContext))

			var agent1 Agent
			err = query.Where("name", "goravel").First(&agent1)
			assert.Nil(t, err)
			assert.True(t, agent1.ID > 0)

			// Test MigrateFreshCommand with --seed flag and no seeders specified
			mockContext = &consolemocks.Context{}
			mockArtisan = &consolemocks.Artisan{}
			mockContext.On("OptionBool", "seed").Return(true).Once()
			mockContext.On("OptionSlice", "seeder").Return([]string{}).Once()
			mockArtisan.On("Call", "db:seed").Return(nil).Once()
			migrateFreshCommand = NewMigrateFreshCommand(mockConfig, mockArtisan)
			assert.Nil(t, migrateFreshCommand.Handle(mockContext))

			var agent2 Agent
			err = query.Where("name", "goravel").First(&agent2)
			assert.Nil(t, err)
			assert.True(t, agent2.ID > 0)

			removeMigrations()
		})
	}
}
