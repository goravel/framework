package grammars

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrefixArray(t *testing.T) {
	values := []string{"a", "b", "c"}
	assert.Equal(t, []string{"prefix a", "prefix b", "prefix c"}, prefixArray("prefix", values))
}
