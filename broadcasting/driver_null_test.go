package broadcasting

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/broadcasting"
)

func TestNullDriver_Broadcast(t *testing.T) {
	driver := NewNullDriver()

	err := driver.Broadcast(
		[]broadcasting.Channel{{Name: "test-channel"}},
		"test.event",
		map[string]any{"key": "value"},
	)
	assert.NoError(t, err)
}
