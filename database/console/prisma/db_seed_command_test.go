package prisma

import (
	"testing"

	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/stretchr/testify/assert"
)

func TestDBSeedCommand(t *testing.T) {
	dbsc := NewDBSeedCommand()
	mockCtx := &consolemocks.Context{}

	// init prisma before any test
	handleInitPrisma(mockCtx, t)

	// fill schema.prisma with data
	fillPrismaSchema()

	// test on user model
	mockCtx.On("Argument", 0).Return("")
	assert.Nil(t, dbsc.Handle(mockCtx))

	// remove prisma after all test
	removePrisma()
}
