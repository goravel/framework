package console

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/config"
	"github.com/goravel/framework/console"
	"github.com/goravel/framework/contracts"
	console2 "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/file"
	testingfile "github.com/goravel/framework/testing/file"
)

func TestJobMakeCommand(t *testing.T) {
	err := testingfile.CreateEnv()
	assert.Nil(t, err)

	configApp := config.ServiceProvider{}
	configApp.Register()

	facadesConfig := facades.Config
	facadesConfig.Add("app", map[string]any{
		"providers": []contracts.ServiceProvider{},
	})

	instance := console.NewApplication()
	instance.Register([]console2.Command{
		&JobMakeCommand{},
	})

	assert.NotPanics(t, func() {
		instance.Call("make:job GoravelJob")
	})

	assert.True(t, file.Exists("app/jobs/goravel_job.go"))
	assert.True(t, file.Remove("app"))
	err = os.Remove(".env")
	assert.Nil(t, err)
}
