package console

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStubsView(t *testing.T) {
	stubs := Stubs{}
	result := stubs.View()

	assert.NotEmpty(t, result)
	assert.Contains(t, result, "DummyPathName")
	assert.Contains(t, result, "DummyPathDefinition")
	assert.Contains(t, result, "DummyViewName")
	assert.Contains(t, result, "{{ define")
	assert.Contains(t, result, "{{ end }}")
}
