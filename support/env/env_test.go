package env

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMainPath(t *testing.T) {
	assert.Equal(t, "github.com/goravel/framework", MainPath())
}
