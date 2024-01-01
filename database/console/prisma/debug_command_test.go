package prisma

import (
	"testing"

	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/stretchr/testify/assert"
)

func TestDebugCommand(t *testing.T) {
	debugCmd := NewDebugCommand()
	mockCtx := &consolemocks.Context{}

	// init prisma
	handleInitPrisma(mockCtx, t)

	// check debug info
	assert.Nil(t, debugCmd.Handle(mockCtx))

	// remove prisma directory after tests
	removePrisma()
}
