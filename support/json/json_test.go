package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshal(t *testing.T) {
	json, err := Marshal(map[string]int{"a": 1})
	assert.Equal(t, []byte(`{"a":1}`), json)
	assert.Nil(t, err)
}

func TestUnmarshal(t *testing.T) {
	var m map[string]int
	err := Unmarshal([]byte(`{"a":1}`), &m)
	assert.Equal(t, map[string]int{"a": 1}, m)
	assert.Nil(t, err)
}

func TestMarshalString(t *testing.T) {
	json, err := MarshalString(map[string]int{"a": 1})
	assert.Equal(t, `{"a":1}`, json)
	assert.Nil(t, err)
}

func TestUnmarshalString(t *testing.T) {
	var m map[string]int
	err := UnmarshalString(`{"a":1}`, &m)
	assert.Equal(t, map[string]int{"a": 1}, m)
	assert.Nil(t, err)
}
