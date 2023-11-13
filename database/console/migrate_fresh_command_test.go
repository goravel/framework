package console

import (
	"testing"

	"github.com/stretchr/testify/assert"

	ormcontract "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/gorm"
	configmock "github.com/goravel/framework/mocks/config"
	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/env"
)

func TestMigrateFreshCommand(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	var (
		mockConfig *configmock.Config
		query      ormcontract.Query
	)

	if err := testDatabaseDocker.Fresh(); err != nil {
		t.Fatal(err)
	}

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
				docker := gorm.NewSqliteDocker("goravel")
				query, err = docker.New()
				assert.Nil(t, err)
				mockConfig = docker.MockConfig
				createSqliteMigrations()
			},
		},
		{
			name: "mysql",
			setup: func() {
				var err error
				docker := gorm.NewMysqlDocker(testDatabaseDocker)
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
				docker := gorm.NewPostgresqlDocker(testDatabaseDocker)
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
				docker := gorm.NewSqlserverDocker(testDatabaseDocker)
				query, err = docker.New()
				assert.Nil(t, err)
				mockConfig = docker.MockConfig
				createSqlserverMigrations()
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
