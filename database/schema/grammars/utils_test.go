package grammars

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/database/schema"
	mocksschema "github.com/goravel/framework/mocks/database/schema"
)

func TestGetCommandByName(t *testing.T) {
	commands := []*schema.Command{
		{Name: "create"},
		{Name: "update"},
		{Name: "delete"},
	}

	// Test case: Command exists
	result := getCommandByName(commands, "update")
	assert.NotNil(t, result)
	assert.Equal(t, "update", result.Name)

	// Test case: Command does not exist
	result = getCommandByName(commands, "drop")
	assert.Nil(t, result)
}

func TestGetDefaultValue(t *testing.T) {
	def := true
	assert.Equal(t, "'1'", getDefaultValue(def))

	def = false
	assert.Equal(t, "'0'", getDefaultValue(def))

	defInt := 123
	assert.Equal(t, "'123'", getDefaultValue(defInt))

	defString := "abc"
	assert.Equal(t, "'abc'", getDefaultValue(defString))
}

func TestGetType(t *testing.T) {
	// valid type
	mockColumn := mocksschema.NewColumnDefinition(t)
	mockColumn.EXPECT().GetType().Return("string").Once()

	mockGrammar := mocksschema.NewGrammar(t)
	mockGrammar.EXPECT().TypeString(mockColumn).Return("varchar").Once()

	assert.Equal(t, "varchar", getType(mockGrammar, mockColumn))

	// invalid type
	mockColumn1 := mocksschema.NewColumnDefinition(t)
	mockColumn1.EXPECT().GetType().Return("invalid").Once()

	mockGrammar1 := mocksschema.NewGrammar(t)

	assert.Empty(t, getType(mockGrammar1, mockColumn1))
}
