package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mocksschema "github.com/goravel/framework/mocks/database/schema"
)

func TestColumnDefaultValue(t *testing.T) {
	def := true
	assert.Equal(t, "'1'", ColumnDefaultValue(def))

	def = false
	assert.Equal(t, "'0'", ColumnDefaultValue(def))

	defInt := 123
	assert.Equal(t, "'123'", ColumnDefaultValue(defInt))

	defString := "abc"
	assert.Equal(t, "'abc'", ColumnDefaultValue(defString))

	defExpression := Expression("abc")
	assert.Equal(t, "abc", ColumnDefaultValue(defExpression))
}

func TestColumnType(t *testing.T) {
	// valid type
	mockColumn := mocksschema.NewColumnDefinition(t)
	mockColumn.EXPECT().GetType().Return("string").Once()

	mockGrammar := mocksschema.NewGrammar(t)
	mockGrammar.EXPECT().TypeString(mockColumn).Return("varchar").Once()

	assert.Equal(t, "varchar", ColumnType(mockGrammar, mockColumn))

	// invalid type
	mockColumn1 := mocksschema.NewColumnDefinition(t)
	mockColumn1.EXPECT().GetType().Return("invalid").Once()

	mockGrammar1 := mocksschema.NewGrammar(t)

	assert.Empty(t, ColumnType(mockGrammar1, mockColumn1))
}
