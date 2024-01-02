package prisma

import (
	"testing"

	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/stretchr/testify/assert"
)

func TestGenerateCommand(t *testing.T) {
	ctx := &consolemocks.Context{}
	genCmd := NewGenerateCommand()

	// init a prisma project before tests
	handleInitPrisma(ctx, t)
	defer removePrisma()

	// fill schema.prisma with data
	fillPrismaSchema()

	// if database instance not running this passes
	assert.Nil(t, genCmd.Handle(ctx))
}
