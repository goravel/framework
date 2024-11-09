package grammars

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/database/schema"
	mocksschema "github.com/goravel/framework/mocks/database/schema"
)

func TestGetColumns(t *testing.T) {
	mockColumn1 := mocksschema.NewColumnDefinition(t)
	mockColumn1.EXPECT().GetName().Return("id").Once()
	mockColumn1.EXPECT().GetType().Return("string").Once()

	mockColumn2 := mocksschema.NewColumnDefinition(t)
	mockColumn2.EXPECT().GetName().Return("name").Once()
	mockColumn2.EXPECT().GetType().Return("string").Once()

	mockBlueprint := mocksschema.NewBlueprint(t)
	mockBlueprint.EXPECT().GetAddedColumns().Return([]schema.ColumnDefinition{
		mockColumn1, mockColumn2,
	}).Once()

	mockGrammar := mocksschema.NewGrammar(t)
	mockGrammar.EXPECT().GetModifiers().Return([]func(schema.Blueprint, schema.ColumnDefinition) string{}).Twice()
	mockGrammar.EXPECT().TypeString(mockColumn1).Return("varchar(100)").Once()
	mockGrammar.EXPECT().TypeString(mockColumn2).Return("varchar").Once()

	assert.Equal(t, []string{"id varchar(100)", "name varchar"}, getColumns(mockGrammar, mockBlueprint))
}

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

func TestPrefixArray(t *testing.T) {
	values := []string{"a", "b", "c"}
	assert.Equal(t, []string{"prefix a", "prefix b", "prefix c"}, prefixArray("prefix", values))
}
