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

func TestMigrateRefreshCommand(t *testing.T) {
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

			mockArtisan := &consolemocks.Artisan{}
			mockContext := &consolemocks.Context{}
			mockContext.On("Option", "step").Return("").Once()

			migrateCommand := NewMigrateCommand(mockConfig)
			assert.Nil(t, migrateCommand.Handle(mockContext))

			// Test MigrateRefreshCommand without --seed flag
			mockContext.On("OptionBool", "seed").Return(false).Once()
			migrateRefreshCommand := NewMigrateRefreshCommand(mockConfig, mockArtisan)
			assert.Nil(t, migrateRefreshCommand.Handle(mockContext))

			var agent Agent
			err := query.Where("name", "goravel").First(&agent)
			assert.Nil(t, err)
			assert.True(t, agent.ID > 0)

			mockArtisan = &consolemocks.Artisan{}
			mockContext = &consolemocks.Context{}
			mockContext.On("Option", "step").Return("5").Once()

			migrateCommand = NewMigrateCommand(mockConfig)
			assert.Nil(t, migrateCommand.Handle(mockContext))

			// Test MigrateRefreshCommand with --seed flag and --seeder specified
			mockContext.On("OptionBool", "seed").Return(true).Once()
			mockContext.On("OptionSlice", "seeder").Return([]string{"UserSeeder"}).Once()
			mockArtisan.On("Call", "db:seed --seeder UserSeeder").Return(nil).Once()
			migrateRefreshCommand = NewMigrateRefreshCommand(mockConfig, mockArtisan)
			assert.Nil(t, migrateRefreshCommand.Handle(mockContext))

			mockArtisan = &consolemocks.Artisan{}
			mockContext = &consolemocks.Context{}

			// Test MigrateRefreshCommand with --seed flag and no --seeder specified
			mockContext.On("Option", "step").Return("").Once()
			mockContext.On("OptionBool", "seed").Return(true).Once()
			mockContext.On("OptionSlice", "seeder").Return([]string{}).Once()
			mockArtisan.On("Call", "db:seed").Return(nil).Once()
			migrateRefreshCommand = NewMigrateRefreshCommand(mockConfig, mockArtisan)
			assert.Nil(t, migrateRefreshCommand.Handle(mockContext))

			var agent1 Agent
			err = query.Where("name", "goravel").First(&agent1)
			assert.Nil(t, err)
			assert.True(t, agent1.ID > 0)

			mockContext.AssertExpectations(t)
			removeMigrations()

		})
	}
}
