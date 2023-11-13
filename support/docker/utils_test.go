package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	_, err := Run("ls")
	assert.Nil(t, err)
}
