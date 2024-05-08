package grammars

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestPrefixArray(t *testing.T) {
	values := []string{"a", "b", "c"}
	assert.Equal(t, []string{"prefix a", "prefix b", "prefix c"}, prefixArray("prefix", values))
}

func TestQuoteString(t *testing.T) {
	values := []string{"a", "b", "c"}
	assert.Equal(t, []string{"'a'", "'b'", "'c'"}, quoteString(values))
}
