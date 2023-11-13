package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/support/env"
)

func TestRun(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	_, err := Run("ls")
	assert.Nil(t, err)
}
