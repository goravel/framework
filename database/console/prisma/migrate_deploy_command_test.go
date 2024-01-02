package prisma

import (
	"testing"

	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/stretchr/testify/assert"
)

func TestMigrateDeployCommand(t *testing.T) {
	ctx := &consolemocks.Context{}
	mdc := NewMigrateDeployCommand()

	// init prisma
	handleInitPrisma(ctx, t)
	defer removePrisma()

	// init prisma schema
	fillPrismaSchema()

	// no args
	ctx.On("Argument", 0).Return("").Once()
	assert.Nil(t, mdc.Handle(ctx))
}
