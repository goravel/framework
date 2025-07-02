package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCookies(t *testing.T) {
	cookies := Cookies(map[string]string{
		"name":  "value",
		"name2": "value2",
	})

	assert.Equal(t, 2, len(cookies))
	assert.Equal(t, "name", cookies[0].Name)
	assert.Equal(t, "value", cookies[0].Value)
	assert.Equal(t, "name2", cookies[1].Name)
	assert.Equal(t, "value2", cookies[1].Value)
}
