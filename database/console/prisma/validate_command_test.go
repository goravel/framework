package prisma

import (
	"testing"

	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/stretchr/testify/assert"
)

func TestValidateCommand(t *testing.T) {
	ctx := &consolemocks.Context{}
	mdc := NewValidateCommand()

	// init prisma
	handleInitPrisma(ctx, t)
	defer removePrisma()

	// create prisma schema
	fillPrismaSchema()

	// no args
	ctx.On("Argument", 0).Return("").Once()
	assert.Nil(t, mdc.Handle(ctx))
}
