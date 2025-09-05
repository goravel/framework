package options

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFacade(t *testing.T) {
	option := Facade("Auth")
	options := make(map[string]any)
	option(options)

	assert.Equal(t, map[string]any{"facade": "Auth"}, options)
}

func TestForce(t *testing.T) {
	option := Force(true)
	options := make(map[string]any)
	option(options)

	assert.Equal(t, map[string]any{"force": true}, options)
}
