package process

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/support/env"
)

func TestRun(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skip test")
	}

	_, err := Run("ls")
	assert.Nil(t, err)
}
