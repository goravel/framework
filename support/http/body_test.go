package http

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBodyImpl_SetField(t *testing.T) {
	body := NewBody()

	body.SetField("name", "Alice").SetField("age", 25)

	assert.Equal(t, "Alice", body.GetField("name"))
	assert.Equal(t, 25, body.GetField("age"))
}

func TestBodyImpl_Build_JSONBody(t *testing.T) {
	body := NewBody()
	body.SetField("name", "Alice").SetField("age", 25)

	reader, err := body.Build()
	assert.NoError(t, err)

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(reader)
	assert.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(buf.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, "Alice", result["name"])
	assert.Equal(t, float64(25), result["age"])
}

func TestBodyImpl_Build_ErrorHandling(t *testing.T) {
	body := NewBody()
	body.SetField("name", "Alice").SetFile("profile", "/invalid/filepath")

	_, err := body.Build()
	assert.Error(t, err, "Expected error for invalid file path")
}
