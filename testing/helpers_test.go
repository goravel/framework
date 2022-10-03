package testing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunInTest(t *testing.T) {
	assert.True(t, RunInTest())
}
