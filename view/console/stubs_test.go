package console

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStubsView(t *testing.T) {
	stubs := Stubs{}
	result := stubs.View()

	assert.NotEmpty(t, result)
	assert.Contains(t, result, "DummyDefinition")
	assert.Contains(t, result, "<h1>Welcome</h1>")
	assert.Contains(t, result, "{{ define")
	assert.Contains(t, result, "{{ end }}")
}
