package env

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPackageName(t *testing.T) {
	assert.Equal(t, "github.com/goravel/framework", PackageName())
}
