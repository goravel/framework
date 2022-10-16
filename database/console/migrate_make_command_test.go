package console

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/config"
	"github.com/goravel/framework/console"
	contractconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/testing/file"
)

func TestMigrateMakeCommand(t *testing.T) {
	configServiceProvider := config.ServiceProvider{}
	configServiceProvider.Register()

	consoleApp := console.Application{}
	instance := consoleApp.Init()
	instance.Register([]contractconsole.Command{
		&MigrateMakeCommand{},
	})

	assert.NotPanics(t, func() {
		instance.Call("make:migration create_users_table")
	})
	assert.NotPanics(t, func() {
		instance.Call("make:migration add_avatar_to_users_table")
	})

	now := time.Now().Format("20060102150405")
	createUpFile := "database/migrations/" + now + "_create_users_table.up.sql"
	createDownFile := "database/migrations/" + now + "_create_users_table.down.sql"

	assert.FileExists(t, createUpFile)
	assert.FileExists(t, createDownFile)

	assert.Equal(t, 9, file.GetLineNum(createUpFile))
	assert.Equal(t, 2, file.GetLineNum(createDownFile))

	updateUpFile := "database/migrations/" + now + "_add_avatar_to_users_table.up.sql"
	updateDownFile := "database/migrations/" + now + "_add_avatar_to_users_table.down.sql"

	assert.FileExists(t, updateUpFile)
	assert.FileExists(t, updateDownFile)

	assert.Equal(t, 2, file.GetLineNum(updateUpFile))
	assert.Equal(t, 2, file.GetLineNum(updateDownFile))

	err := os.RemoveAll("database")
	assert.Nil(t, err)
}
