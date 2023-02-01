package console

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

var testCommand = 0

func TestRun(t *testing.T) {
	cli := NewApplication()
	cli.Register([]console.Command{
		&TestCommand{},
	})

	cli.Call("test")
	assert.Equal(t, 1, testCommand)
}

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
	testCommand++

	return nil
}
