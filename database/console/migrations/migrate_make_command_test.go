package migrations

import (
	"github.com/goravel/framework/config"
	"github.com/goravel/framework/console"
	"github.com/goravel/framework/console/support"
	support2 "github.com/goravel/framework/support"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestMigrateMakeCommand(t *testing.T) {
	configServiceProvider := config.ServiceProvider{}
	configServiceProvider.Register()

	consoleApp := console.Application{}
	consoleApp.Init().Register([]support.Command{
		MigrateMakeCommand{},
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

	assert.Equal(t, 8, support2.GetLineNum(createUpFile))
	assert.Equal(t, 1, support2.GetLineNum(createDownFile))

	updateUpFile := "database/migrations/" + now + "_add_avatar_to_users_table.up.sql"
	updateDownFile := "database/migrations/" + now + "_add_avatar_to_users_table.down.sql"

	assert.FileExists(t, updateUpFile)
	assert.FileExists(t, updateDownFile)

	assert.Equal(t, 1, support2.GetLineNum(updateUpFile))
	assert.Equal(t, 1, support2.GetLineNum(updateDownFile))

	err := os.RemoveAll("database")
	assert.Nil(t, err)
}
