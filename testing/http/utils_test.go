package http

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCookies(t *testing.T) {
	cookies := Cookies(map[string]string{
		"name":  "value",
		"name2": "value2",
	})

	assert.Equal(t, 2, len(cookies))
	assert.ElementsMatch(t, cookies, []*http.Cookie{
		{
			Name:  "name",
			Value: "value",
		},
		{
			Name:  "name2",
			Value: "value2",
		},
	})
}
