package env

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModuleName(t *testing.T) {
	assert.Equal(t, "github.com/goravel/framework", PackageName())
}
