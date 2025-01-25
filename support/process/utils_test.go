package process

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidPort(t *testing.T) {
	assert.True(t, ValidPort() > 0)
}
