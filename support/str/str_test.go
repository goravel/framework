package str

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRandom(t *testing.T) {
	assert.Len(t, Random(10), 10)
}
