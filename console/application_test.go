package console

import (
	"testing"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/stretchr/testify/assert"
)

type TestCommand struct {
}

func (receiver *TestCommand) Signature() string {
	return "test"
}

func (receiver *TestCommand) Description() string {
	return "Test command"
}

func (receiver *TestCommand) Extend() command.Extend {
	return command.Extend{}
}

func (receiver *TestCommand) Handle(ctx console.Context) error {
	return nil
}

func TestInit(t *testing.T) {
	assert.NotPanics(t, func() {
		app := Application{}
		app.Init()
	})
}

func TestRun(t *testing.T) {
	app := Application{}
	cli := app.Init()
	cli.Register([]console.Command{
		&TestCommand{},
	})

	assert.NotPanics(t, func() {
		cli.Call("test")
	})
}
