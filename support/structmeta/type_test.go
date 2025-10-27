package structmeta

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTagMetadata(t *testing.T) {
	tag := reflect.StructTag(`json:"id,omitempty" validate:"min=1,max=10"`)
	meta := NewTagMetadata(tag)

	assert.True(t, meta.HasKey("json"))
	assert.True(t, meta.HasKey("validate"))
	assert.False(t, meta.HasKey("nonexistent"))

	assert.Equal(t, "id,omitempty", meta.Get("json"))
	assert.Equal(t, []string{"id", "omitempty"}, meta.GetParts("json"))

	assert.Equal(t, "min=1,max=10", meta.Get("validate"))
	assert.Equal(t, []string{"min=1", "max=10"}, meta.GetParts("validate"))
}
