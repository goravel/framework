package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrors_One(t *testing.T) {
	t.Run("empty errors", func(t *testing.T) {
		errors := NewErrors()
		assert.Equal(t, "", errors.One())
	})

	t.Run("returns first error", func(t *testing.T) {
		errors := NewErrors()
		errors.Add("a", "required", "The a field is required.")
		assert.Equal(t, "The a field is required.", errors.One())
	})

	t.Run("returns first error for specific field", func(t *testing.T) {
		errors := NewErrors()
		errors.Add("a", "required", "The a field is required.")
		errors.Add("b", "required", "The b field is required.")
		assert.Equal(t, "The a field is required.", errors.One("a"))
		assert.Equal(t, "The b field is required.", errors.One("b"))
	})
}

func TestErrors_Get(t *testing.T) {
	t.Run("empty errors", func(t *testing.T) {
		errors := NewErrors()
		assert.Empty(t, errors.Get("a"))
	})

	t.Run("returns field errors", func(t *testing.T) {
		errors := NewErrors()
		errors.Add("a", "required", "The a field is required.")
		errors.Add("b", "required", "The b field is required.")
		assert.Equal(t, map[string]string{"required": "The a field is required."}, errors.Get("a"))
		assert.Equal(t, map[string]string{"required": "The b field is required."}, errors.Get("b"))
	})
}

func TestErrors_All(t *testing.T) {
	t.Run("empty errors", func(t *testing.T) {
		errors := NewErrors()
		assert.Empty(t, errors.All())
	})

	t.Run("returns all errors", func(t *testing.T) {
		errors := NewErrors()
		errors.Add("a", "required", "The a field is required.")
		errors.Add("b", "required", "The b field is required.")
		assert.Equal(t, map[string]map[string]string{
			"a": {"required": "The a field is required."},
			"b": {"required": "The b field is required."},
		}, errors.All())
	})
}

func TestErrors_Has(t *testing.T) {
	t.Run("empty errors", func(t *testing.T) {
		errors := NewErrors()
		assert.False(t, errors.Has("a"))
	})

	t.Run("has field with errors", func(t *testing.T) {
		errors := NewErrors()
		errors.Add("a", "required", "The a field is required.")
		assert.True(t, errors.Has("a"))
		assert.False(t, errors.Has("b"))
	})
}

func TestErrors_IsEmpty(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		errors := NewErrors()
		assert.True(t, errors.IsEmpty())
	})

	t.Run("not empty", func(t *testing.T) {
		errors := NewErrors()
		errors.Add("a", "required", "The a field is required.")
		assert.False(t, errors.IsEmpty())
	})
}
