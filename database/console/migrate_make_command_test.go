package console

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/config"
	"github.com/goravel/framework/console"
	console2 "github.com/goravel/framework/contracts/console"
	goraveltesting "github.com/goravel/framework/testing"
)

func TestMigrateMakeCommand(t *testing.T) {
	configServiceProvider := config.ServiceProvider{}
	configServiceProvider.Register()

	consoleApp := console.Application{}
	consoleApp.Init().Register([]console2.Command{
		&MigrateMakeCommand{},
	})

	assert.NotPanics(t, func() {
		consoleApp.Call("make:migration create_users_table")
	})
	assert.NotPanics(t, func() {
		consoleApp.Call("make:migration add_avatar_to_users_table")
	})

	now := time.Now().Format("20060102150405")
	createUpFile := "database/migrations/" + now + "_create_users_table.up.sql"
	createDownFile := "database/migrations/" + now + "_create_users_table.down.sql"

	assert.FileExists(t, createUpFile)
	assert.FileExists(t, createDownFile)

	assert.Equal(t, 9, goraveltesting.GetLineNum(createUpFile))
	assert.Equal(t, 2, goraveltesting.GetLineNum(createDownFile))

	updateUpFile := "database/migrations/" + now + "_add_avatar_to_users_table.up.sql"
	updateDownFile := "database/migrations/" + now + "_add_avatar_to_users_table.down.sql"

	assert.FileExists(t, updateUpFile)
	assert.FileExists(t, updateDownFile)

	assert.Equal(t, 2, goraveltesting.GetLineNum(updateUpFile))
	assert.Equal(t, 2, goraveltesting.GetLineNum(updateDownFile))

	err := os.RemoveAll("database")
	assert.Nil(t, err)
}
