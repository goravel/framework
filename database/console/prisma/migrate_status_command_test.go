package prisma

import (
	"testing"

	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/stretchr/testify/assert"
)

func TestMigrateStatusCommand(t *testing.T) {
	ctx := &consolemocks.Context{}
	mdc := NewMigrateStatusCommand()

	// init prisma
	handleInitPrisma(ctx, t)
	defer removePrisma()

	// fill prisma schema
	fillPrismaSchema()

	// migrate dev to get database tables ready
	migrateDev(ctx)

	// no args
	ctx.On("Argument", 0).Return("").Once()
	assert.Nil(t, mdc.Handle(ctx))
}
