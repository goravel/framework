package prisma

import (
	"testing"

	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/stretchr/testify/assert"
)

func TestFormatCommand(t *testing.T) {
	fmtCmd := NewFormatCommand()
	ctx := &consolemocks.Context{}

	// init prisma project
	handleInitPrisma(ctx, t)

	// fill schema.prisma with data
	fillPrismaSchema()

	// formatting the prisma schema file has no error
	assert.NoError(t, fmtCmd.Handle(ctx))

	// remove prisma directory
	removePrisma()
}
