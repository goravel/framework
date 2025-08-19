package file

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLineNum(t *testing.T) {
	assert.Equal(t, 33, GetLineNum("file.go"))
}
