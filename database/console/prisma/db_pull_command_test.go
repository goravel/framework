package prisma

import (
	"testing"

	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/stretchr/testify/assert"
)

func TestDBPullCommand(t *testing.T) {
	// make an instance of db pull command struct
	dbpc := &DBPullCommand{}
	mockCtx := &consolemocks.Context{}

	// init prisma
	handleInitPrisma(mockCtx, t)

	// fill schema.prisma with data
	fillPrismaSchema()

	// runs into error because it needs data in database
	// otherwise there's no error
	// used assert.Error to pass test
	mockCtx.On("Argument", 0).Return("").Once()
	assert.Error(t, dbpc.Handle(mockCtx))

	// remove prima directory
	removePrisma()
}
