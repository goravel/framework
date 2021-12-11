package console

import (
	"github.com/goravel/framework/console/support"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"testing"
)

type TestCommand struct {
}

func (receiver TestCommand) Signature() string {
	return "test"
}

func (receiver TestCommand) Description() string {
	return "Test command"
}

func (receiver TestCommand) Flags() []cli.Flag {
	var flags []cli.Flag

	return flags
}

func (receiver TestCommand) Subcommands() []*cli.Command {
	var subcommands []*cli.Command

	return subcommands
}

func (receiver TestCommand) Handle(c *cli.Context) error {

	return nil
}

func TestRegister(t *testing.T) {
	app := Application{}
	app.Register([]support.Command{
		TestCommand{},
	})

	assert.Equal(t, len(app.GetInstance().Commands), 1)
}

func TestRun(t *testing.T) {
	app := Application{}
	app.Register([]support.Command{
		TestCommand{},
	})

	assert.Panics(t, func() {
		app.run([]string{
			"./main",
			"artisan",
			"test",
		})
	})
}
