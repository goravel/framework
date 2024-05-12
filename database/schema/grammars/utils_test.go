package grammars

import (
	"testing"

	"github.com/stretchr/testify/assert"

	schemacontract "github.com/goravel/framework/contracts/database/schema"
	mockschema "github.com/goravel/framework/mocks/database/schema"
)

func TestGetColumns(t *testing.T) {
	mockColumn1 := &mockschema.ColumnDefinition{}
	mockColumn1.EXPECT().GetName().Return("id").Once()
	mockColumn1.EXPECT().GetType().Return("string").Once()
	mockColumn2 := &mockschema.ColumnDefinition{}
	mockColumn2.EXPECT().GetName().Return("name").Once()
	mockColumn2.EXPECT().GetType().Return("string").Once()
	mockBlueprint := &mockschema.Blueprint{}
	mockBlueprint.EXPECT().GetAddedColumns().Return([]schemacontract.ColumnDefinition{
		mockColumn1, mockColumn2,
	}).Once()
	mockGrammar := &mockschema.Grammar{}
	mockGrammar.EXPECT().GetModifiers().Return([]func(schemacontract.Blueprint, schemacontract.ColumnDefinition) string{}).Twice()
	mockGrammar.EXPECT().TypeString(mockColumn1).Return("varchar(100)").Once()
	mockGrammar.EXPECT().TypeString(mockColumn2).Return("varchar").Once()

	assert.Equal(t, []string{"id varchar(100)", "name varchar"}, getColumns(mockGrammar, mockBlueprint))

	mockColumn1.AssertExpectations(t)
	mockColumn2.AssertExpectations(t)
	mockBlueprint.AssertExpectations(t)
	mockGrammar.AssertExpectations(t)
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
	mockColumn := &mockschema.ColumnDefinition{}
	mockColumn.EXPECT().GetType().Return("text").Once()
	mockGrammar := &mockschema.Grammar{}
	mockGrammar.EXPECT().TypeText(mockColumn).Return("text").Once()

	assert.Equal(t, "text", getType(mockGrammar, mockColumn))

	mockColumn.AssertExpectations(t)
	mockGrammar.AssertExpectations(t)

	// invalid type
	mockColumn = &mockschema.ColumnDefinition{}
	mockColumn.EXPECT().GetType().Return("invalid").Once()
	mockGrammar = &mockschema.Grammar{}

	assert.Empty(t, getType(mockGrammar, mockColumn))

	mockColumn.AssertExpectations(t)
	mockGrammar.AssertExpectations(t)
}

func TestPrefixArray(t *testing.T) {
	values := []string{"a", "b", "c"}
	assert.Equal(t, []string{"prefix a", "prefix b", "prefix c"}, prefixArray("prefix", values))
}

func TestQuoteString(t *testing.T) {
	values := []string{"a", "b", "c"}
	assert.Equal(t, []string{"'a'", "'b'", "'c'"}, quoteString(values))
}
