package console

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGuess(t *testing.T) {
	table, create := TableGuesser{}.Guess("create_users_table")
	assert.Equal(t, table, "users")
	assert.Equal(t, create, true)

	table, create = TableGuesser{}.Guess("add_avatar_to_users_table")
	assert.Equal(t, table, "users")
	assert.Equal(t, create, false)
}
