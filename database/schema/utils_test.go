package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mocksdriver "github.com/goravel/framework/mocks/database/driver"
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
	mockColumn := mocksdriver.NewColumnDefinition(t)
	mockColumn.EXPECT().GetType().Return("string").Once()

	mockGrammar := mocksdriver.NewGrammar(t)
	mockGrammar.EXPECT().TypeString(mockColumn).Return("varchar").Once()

	assert.Equal(t, "varchar", ColumnType(mockGrammar, mockColumn))

	// invalid type
	mockColumn1 := mocksdriver.NewColumnDefinition(t)
	mockColumn1.EXPECT().GetType().Return("invalid").Twice()

	mockGrammar1 := mocksdriver.NewGrammar(t)

	assert.Equal(t, "invalid", ColumnType(mockGrammar1, mockColumn1))
}
