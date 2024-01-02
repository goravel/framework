package prisma

import (
	"testing"

	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/stretchr/testify/assert"
)

func TestVersionCommand(t *testing.T) {
	ctx := &consolemocks.Context{}
	mdc := NewVersionCommand()

	// no args
	ctx.On("Argument", 0).Return("").Once()
	assert.Nil(t, mdc.Handle(ctx))
}
