package console

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	"github.com/goravel/framework/support/file"
)

var databaseKernel = `package database

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/database/seeder"
)

type Kernel struct {
}

func (kernel Kernel) Migrations() []schema.Migration {
	return []schema.Migration{}
}

func (kernel Kernel) Seeders() []seeder.Seeder {
	return []seeder.Seeder{}
}`

func TestSeederMakeCommand(t *testing.T) {
	mockApp := mocksfoundation.NewApplication(t)
	seederMakeCommand := &SeederMakeCommand{app: mockApp}

	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Argument(0).Return("").Once()
	mockContext.EXPECT().Ask("Enter the seeder name", mock.Anything).Return("", errors.New("the seeder name cannot be empty")).Once()
	mockContext.EXPECT().Error("the seeder name cannot be empty").Once()
	assert.Nil(t, seederMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("UserSeeder").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockApp.EXPECT().DatabasePath("kernel.go").Return("database/kernel.go").Once()
	mockContext.EXPECT().Success("Seeder created successfully").Once()
	mockContext.EXPECT().Success("Seeder registered successfully").Once()
	assert.NoError(t, file.PutContent("database/kernel.go", databaseKernel))
	assert.NoError(t, seederMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("database/seeders/user_seeder.go"))
	assert.True(t, file.Contain("database/seeders/user_seeder.go", "package seeders"))
	assert.True(t, file.Contain("database/seeders/user_seeder.go", "type UserSeeder struct"))
	assert.True(t, file.Contain("database/kernel.go", "database/seeders"))
	assert.True(t, file.Contain("database/kernel.go", "&seeders.UserSeeder{}"))

	mockContext.EXPECT().Argument(0).Return("UserSeeder").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Error("the seeder already exists. Use the --force or -f flag to overwrite").Once()
	assert.NoError(t, seederMakeCommand.Handle(mockContext))
	assert.NoError(t, file.Remove("database"))

	mockContext.EXPECT().Argument(0).Return("subdir/DemoSeeder").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Seeder created successfully").Once()
	mockApp.EXPECT().DatabasePath("kernel.go").Return("database/kernel.go").Once()
	mockContext.EXPECT().Warning(mock.MatchedBy(func(msg string) bool {
		return strings.HasPrefix(msg, "seeder register failed:")
	})).Once()
	assert.NoError(t, seederMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("database/seeders/subdir/demo_seeder.go"))
	assert.True(t, file.Contain("database/seeders/subdir/demo_seeder.go", "package subdir"))
	assert.True(t, file.Contain("database/seeders/subdir/demo_seeder.go", "type DemoSeeder struct"))
	assert.NoError(t, file.Remove("database"))
}

func TestSeederMakeCommand_BootstrapSetup(t *testing.T) {
	var (
		mockContext *mocksconsole.Context
		mockApp     *mocksfoundation.Application
	)

	// Create bootstrap/app.go to trigger IsBootstrapSetup() == true
	bootstrapContent := `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/contracts/database/seeder"
)

func Boot() {
	foundation.Setup().WithSeeders([]seeder.Seeder{}).Run()
}
`
	assert.NoError(t, file.PutContent("bootstrap/app.go", bootstrapContent))

	// Cleanup after test
	defer func() {
		assert.NoError(t, file.Remove("bootstrap"))
		assert.NoError(t, file.Remove("database"))
	}()

	t.Run("Bootstrap setup - successful registration", func(t *testing.T) {
		mockContext = mocksconsole.NewContext(t)
		mockApp = mocksfoundation.NewApplication(t)

		mockContext.EXPECT().Argument(0).Return("UserSeeder").Once()
		mockContext.EXPECT().OptionBool("force").Return(false).Once()
		mockContext.EXPECT().Success("Seeder created successfully").Once()
		mockContext.EXPECT().Success("Seeder registered successfully").Once()

		seederMakeCommand := NewSeederMakeCommand(mockApp)
		err := seederMakeCommand.Handle(mockContext)

		assert.NoError(t, err)
		assert.True(t, file.Exists("database/seeders/user_seeder.go"))
		assert.True(t, file.Contain("database/seeders/user_seeder.go", "package seeders"))
		assert.True(t, file.Contain("database/seeders/user_seeder.go", "type UserSeeder struct"))

		// Verify bootstrap/app.go was updated with the seeder
		bootstrapUpdated, err := file.GetContent("bootstrap/app.go")
		assert.NoError(t, err)
		assert.Contains(t, bootstrapUpdated, "database/seeders")
		assert.Contains(t, bootstrapUpdated, "&seeders.UserSeeder{}")
	})

	t.Run("Bootstrap setup - registration failed", func(t *testing.T) {
		// Reset bootstrap/app.go with invalid syntax
		invalidBootstrapContent := `package bootstrap

import (
	"github.com/goravel/framework/foundation"
)

func Boot() {
	foundation.Setup().
}
`
		assert.NoError(t, file.PutContent("bootstrap/app.go", invalidBootstrapContent))

		mockContext = mocksconsole.NewContext(t)
		mockApp = mocksfoundation.NewApplication(t)

		mockContext.EXPECT().Argument(0).Return("PostSeeder").Once()
		mockContext.EXPECT().OptionBool("force").Return(false).Once()
		mockContext.EXPECT().Success("Seeder created successfully").Once()
		mockContext.EXPECT().Warning(mock.MatchedBy(func(msg string) bool {
			return strings.Contains(msg, "seeder register failed:")
		})).Once()

		seederMakeCommand := NewSeederMakeCommand(mockApp)
		err := seederMakeCommand.Handle(mockContext)

		assert.NoError(t, err)
		assert.True(t, file.Exists("database/seeders/post_seeder.go"))
		assert.True(t, file.Contain("database/seeders/post_seeder.go", "package seeders"))
		assert.True(t, file.Contain("database/seeders/post_seeder.go", "type PostSeeder struct"))
	})
}
