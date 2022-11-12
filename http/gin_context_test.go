package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {
	httpCtx := Background()
	httpCtx.WithValue("Hello", "world")
	httpCtx.WithValue("Hi", "Goravel")
	ctx := httpCtx.Context()
	assert.Equal(t, ctx.Value("Hello").(string), "world")
	assert.Equal(t, ctx.Value("Hi").(string), "Goravel")
}
