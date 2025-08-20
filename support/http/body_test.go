package http

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/errors"
)

func TestBodySetFields(t *testing.T) {
	body := NewBody().SetFields(map[string]any{
		"name": "krishan",
		"age":  22,
		"role": "developer",
	}).SetField("role", "admin")

	assert.Equal(t, "krishan", body.GetField("name"))
	assert.Equal(t, 22, body.GetField("age"))
	assert.Equal(t, "admin", body.GetField("role"))
}

func TestBodySetField(t *testing.T) {
	body := NewBody().
		SetField("name", "krishan").
		SetField("age", 22)

	assert.Equal(t, "krishan", body.GetField("name"))
	assert.Equal(t, 22, body.GetField("age"))
}

func TestBuildJSONBody(t *testing.T) {
	body := NewBody().
		SetField("name", "krishan").
		SetField("age", 22)

	reader, err := body.Build()
	assert.NoError(t, err)

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(reader.Reader())
	assert.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(buf.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, "krishan", result["name"])
	assert.Equal(t, float64(22), result["age"])
}

func TestBuildFormBody(t *testing.T) {
	body := NewBody(BodyTypeForm).
		SetField("name", "krishan").
		SetField("age", 22)

	reader, err := body.Build()
	assert.NoError(t, err)

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(reader.Reader())
	assert.NoError(t, err)

	formData, err := url.ParseQuery(buf.String())
	assert.NoError(t, err)
	assert.Equal(t, "krishan", formData.Get("name"))
	assert.Equal(t, "22", formData.Get("age"))
}

func TestBuildMultipartBody(t *testing.T) {
	file, err := os.CreateTemp("", "example.txt")
	assert.NoError(t, err)
	defer func() {
		_ = file.Close()
		_ = os.Remove(file.Name())
	}()
	_, err = file.WriteString("file content")
	assert.NoError(t, err)

	body := NewBody().
		SetField("name", "krishan").
		SetFile("file", file.Name())

	reader, err := body.Build()
	assert.NoError(t, err)

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(reader.Reader())
	assert.NoError(t, err)

	contentType := reader.ContentType()
	_, params, err := mime.ParseMediaType(contentType)
	assert.NoError(t, err)
	boundary := params["boundary"]
	mr := multipart.NewReader(buf, boundary)
	form, err := mr.ReadForm(1024)
	assert.NoError(t, err)

	assert.NotNil(t, form.Value["name"])
	assert.Greater(t, len(form.Value["name"]), 0)
	assert.Equal(t, "krishan", form.Value["name"][0])

	fileHeaders, ok := form.File["file"]
	assert.True(t, ok)
	fileReader, err := fileHeaders[0].Open()
	assert.NoError(t, err)
	defer errors.Ignore(fileReader.Close)
	fileContent, err := io.ReadAll(fileReader)
	assert.NoError(t, err)
	assert.Equal(t, "file content", string(fileContent))
}

func TestBuildWithError(t *testing.T) {
	body := NewBody().
		SetField("name", "krishan").
		SetFile("profile", "invalid.txt")

	_, err := body.Build()
	assert.Error(t, err, "Expected error for invalid file path")
}

func TestMultipartWithInvalidFilePath(t *testing.T) {
	body := NewBody(BodyTypeMultipart).
		SetField("name", "krishan").
		SetFiles(map[string][]string{"nonexistent": {"nonexistent.txt"}})

	_, err := body.Build()
	assert.Error(t, err, "Expected error for nonexistent file path")
}

func TestBuildFormBodyWithSpecialCharacters(t *testing.T) {
	body := NewBody(BodyTypeForm).
		SetField("name", "krishan").
		SetField("special_chars", "&=+?")

	reader, err := body.Build()
	assert.NoError(t, err)

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(reader.Reader())
	assert.NoError(t, err)

	formData, err := url.ParseQuery(buf.String())
	assert.NoError(t, err)
	assert.Equal(t, "krishan", formData.Get("name"))
	assert.Equal(t, "&=+?", formData.Get("special_chars"))
}
