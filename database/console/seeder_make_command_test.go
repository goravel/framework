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
