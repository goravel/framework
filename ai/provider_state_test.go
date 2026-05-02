package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProviderState(t *testing.T) {
	t.Run("get missing key", func(t *testing.T) {
		state := newProviderState()

		assert.Nil(t, state.Get("missing"))
		assert.Nil(t, state.data)
	})

	t.Run("set and get key", func(t *testing.T) {
		state := newProviderState()

		state.Set("foo", "bar")

		assert.Equal(t, "bar", state.Get("foo"))
		assert.NotNil(t, state.data)
	})

	t.Run("delete last key clears backing map", func(t *testing.T) {
		state := newProviderState()

		state.Set("foo", "bar")
		state.Set("foo", nil)

		assert.Nil(t, state.Get("foo"))
		assert.Nil(t, state.data)
	})

	t.Run("delete key keeps other values", func(t *testing.T) {
		state := newProviderState()

		state.Set("foo", "bar")
		state.Set("baz", 123)
		state.Set("foo", nil)

		assert.Nil(t, state.Get("foo"))
		assert.Equal(t, 123, state.Get("baz"))
		assert.Len(t, state.data, 1)
	})
}
