package console

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/file"
)

func TestSeederMakeCommand(t *testing.T) {
	seederMakeCommand := &SeederMakeCommand{}
	mockContext := &consolemocks.Context{}
	mockContext.On("Argument", 0).Return("").Once()
	mockContext.On("Ask", "Enter the seeder name", mock.Anything).Return("", errors.New("the seeder name cannot be empty")).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, seederMakeCommand.Handle(mockContext))
	}), "the seeder name cannot be empty")

	mockContext.On("Argument", 0).Return("UserSeeder").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Nil(t, seederMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("database/seeders/user_seeder.go"))
	assert.True(t, file.Contain("database/seeders/user_seeder.go", "package seeders"))
	assert.True(t, file.Contain("database/seeders/user_seeder.go", "type UserSeeder struct"))

	mockContext.On("Argument", 0).Return("UserSeeder").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, seederMakeCommand.Handle(mockContext))
	}), "the seeder already exists. Use the --force or -f flag to overwrite")
	assert.Nil(t, file.Remove("database"))

	mockContext.On("Argument", 0).Return("subdir/DemoSeeder").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Nil(t, seederMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("database/seeders/subdir/demo_seeder.go"))
	assert.True(t, file.Contain("database/seeders/subdir/demo_seeder.go", "package subdir"))
	assert.True(t, file.Contain("database/seeders/subdir/demo_seeder.go", "type DemoSeeder struct"))
	assert.Nil(t, file.Remove("database"))
}
