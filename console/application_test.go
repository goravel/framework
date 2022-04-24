package console

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"testing"
)

type TestCommand struct {
}

func (receiver *TestCommand) Signature() string {
	return "test"
}

func (receiver *TestCommand) Description() string {
	return "Test command"
}

func (receiver *TestCommand) Extend() console.CommandExtend {
	return console.CommandExtend{}
}

func (receiver *TestCommand) Handle(c *cli.Context) error {
	return nil
}

func TestInit(t *testing.T) {
	assert.NotPanics(t, func() {
		app := Application{}
		app.Init()
	})
}

func TestRegister(t *testing.T) {
	app := Application{}
	app.Init()
	app.Register([]console.Command{
		&TestCommand{},
	})

	assert.Equal(t, len(app.cli.Commands), 1)
}

func TestRun(t *testing.T) {
	app := Application{}
	app.Init()
	app.Register([]console.Command{
		&TestCommand{},
	})

	assert.NotPanics(t, func() {
		app.Call("test")
	})
}
