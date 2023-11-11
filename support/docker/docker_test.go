package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	assert.Nil(t, Run("ls"))
}
