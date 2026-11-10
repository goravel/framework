package console

type Stubs struct {
}

func (r Stubs) Command() string {
	return `package DummyPackage

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

type DummyCommand struct {
}

// Signature The name and signature of the console command.
func (r *DummyCommand) Signature() string {
	return "DummySignature"
}

// Description The console command description.
func (r *DummyCommand) Description() string {
	return "Command description"
}

// Extend The console command extend.
func (r *DummyCommand) Extend() command.Extend {
	return command.Extend{Category: "app"}
}

// Handle Execute the console command.
func (r *DummyCommand) Handle(ctx console.Context) error {
	
	return nil
}
`
}

func (r Stubs) CommandHook() string {
	return `package DummyPackage

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

// DummyCommandHook runs around the Handle of the commands it is registered for.
type DummyCommandHook struct {
}

// Before runs before Handle. Returning false or an error skips Handle and every After.
func (r *DummyCommandHook) Before(ctx console.Context) (bool, error) {
	return true, nil
}

// After runs after Handle and receives its error. The returned error replaces it.
func (r *DummyCommandHook) After(ctx console.Context, handleErr error) error {
	return handleErr
}

// Flags are added to every matching command. Return nil for none.
func (r *DummyCommandHook) Flags() []command.Flag {
	return nil
}
`
}
