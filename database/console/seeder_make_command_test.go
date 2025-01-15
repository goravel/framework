package console

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/file"
)

func TestSeederMakeCommand(t *testing.T) {
	seederMakeCommand := &SeederMakeCommand{}
	mockContext := &mocksconsole.Context{}
	mockContext.EXPECT().Argument(0).Return("").Once()
	mockContext.EXPECT().Ask("Enter the seeder name", mock.Anything).Return("", errors.New("the seeder name cannot be empty")).Once()
	mockContext.EXPECT().Error("the seeder name cannot be empty").Once()
	assert.Nil(t, seederMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("UserSeeder").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Seeder created successfully").Once()
	assert.Nil(t, seederMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("database/seeders/user_seeder.go"))
	assert.True(t, file.Contain("database/seeders/user_seeder.go", "package seeders"))
	assert.True(t, file.Contain("database/seeders/user_seeder.go", "type UserSeeder struct"))

	mockContext.EXPECT().Argument(0).Return("UserSeeder").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Error("the seeder already exists. Use the --force or -f flag to overwrite").Once()
	assert.Nil(t, seederMakeCommand.Handle(mockContext))
	assert.Nil(t, file.Remove("database"))

	mockContext.EXPECT().Argument(0).Return("subdir/DemoSeeder").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Seeder created successfully").Once()
	assert.Nil(t, seederMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("database/seeders/subdir/demo_seeder.go"))
	assert.True(t, file.Contain("database/seeders/subdir/demo_seeder.go", "package subdir"))
	assert.True(t, file.Contain("database/seeders/subdir/demo_seeder.go", "type DemoSeeder struct"))
	assert.Nil(t, file.Remove("database"))
}
