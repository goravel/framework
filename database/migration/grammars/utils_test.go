package grammars

import (
	"testing"

	"github.com/stretchr/testify/assert"

	contractsmigration "github.com/goravel/framework/contracts/database/migration"
	mockmigration "github.com/goravel/framework/mocks/database/migration"
)

func TestGetColumns(t *testing.T) {
	mockColumn1 := mockmigration.NewColumnDefinition(t)
	mockColumn1.EXPECT().GetName().Return("id").Once()
	mockColumn1.EXPECT().GetType().Return("string").Once()

	mockColumn2 := mockmigration.NewColumnDefinition(t)
	mockColumn2.EXPECT().GetName().Return("name").Once()
	mockColumn2.EXPECT().GetType().Return("string").Once()

	mockBlueprint := mockmigration.NewBlueprint(t)
	mockBlueprint.EXPECT().GetAddedColumns().Return([]contractsmigration.ColumnDefinition{
		mockColumn1, mockColumn2,
	}).Once()

	mockGrammar := mockmigration.NewGrammar(t)
	mockGrammar.EXPECT().GetModifiers().Return([]func(contractsmigration.Blueprint, contractsmigration.ColumnDefinition) string{}).Twice()
	mockGrammar.EXPECT().TypeString(mockColumn1).Return("varchar(100)").Once()
	mockGrammar.EXPECT().TypeString(mockColumn2).Return("varchar").Once()

	assert.Equal(t, []string{"id varchar(100)", "name varchar"}, getColumns(mockGrammar, mockBlueprint))
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
	mockColumn := mockmigration.NewColumnDefinition(t)
	mockColumn.EXPECT().GetType().Return("string").Once()

	mockGrammar := mockmigration.NewGrammar(t)
	mockGrammar.EXPECT().TypeString(mockColumn).Return("varchar").Once()

	assert.Equal(t, "varchar", getType(mockGrammar, mockColumn))

	// invalid type
	mockColumn1 := mockmigration.NewColumnDefinition(t)
	mockColumn1.EXPECT().GetType().Return("invalid").Once()

	mockGrammar1 := mockmigration.NewGrammar(t)

	assert.Empty(t, getType(mockGrammar1, mockColumn1))
}
