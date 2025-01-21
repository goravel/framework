package file

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/testing/file"
)

func TestClientOriginalExtension(t *testing.T) {
	assert.Equal(t, ClientOriginalExtension("logo.png"), "png")
}

func TestContain(t *testing.T) {
	assert.True(t, Contain("../constant.go", "const Version"))
}

func TestCreate(t *testing.T) {
	filePath := path.Join(t.TempDir(), "goravel.txt")
	assert.Nil(t, Create(filePath, `goravel`))
	assert.Equal(t, 1, file.GetLineNum(filePath))
	assert.True(t, Exists(filePath))
}

func TestExists(t *testing.T) {
	assert.True(t, Exists("file.go"))
}

func TestExtension(t *testing.T) {
	extension, err := Extension("file.go")
	assert.Nil(t, err)
	assert.Equal(t, "txt", extension)
}

func TestLastModified(t *testing.T) {
	ti, err := LastModified("../../logo.png", "UTC")
	assert.Nil(t, err)
	assert.NotNil(t, ti)
}

func TestMimeType(t *testing.T) {
	mimeType, err := MimeType("../../logo.png")
	assert.Nil(t, err)
	assert.Equal(t, "image/png", mimeType)
}

func TestRemove(t *testing.T) {
	pwd, _ := os.Getwd()
	filePath := path.Join(pwd, "goravel/goravel.txt")
	assert.Nil(t, Create(filePath, `goravel`))

	assert.Nil(t, Remove(filePath))
	assert.Nil(t, Remove(path.Join(pwd, "goravel")))
}

func TestSize(t *testing.T) {
	size, err := Size("../../logo.png")
	assert.Nil(t, err)
	assert.Equal(t, int64(10853), size)
}

func TestGetContent(t *testing.T) {
	// file not exists
	content, err := GetContent("files.go")
	assert.NotNil(t, err)
	assert.Empty(t, content)
	// get content successfully
	content, err = GetContent("../constant.go")
	assert.Nil(t, err)
	assert.NotNil(t, content)
	assert.Contains(t, content, "const Version")
}

func TestPutContent(t *testing.T) {
	filePath := path.Join(t.TempDir(), "goravel.txt")
	// Create a file and put content
	assert.NoError(t, PutContent(filePath, "goravel"))
	assert.True(t, Contain(filePath, "goravel"))
	assert.Equal(t, 1, file.GetLineNum(filePath))
	// Append content
	assert.NoError(t, PutContent(filePath, "\nframework", WithAppend(true)))
	assert.True(t, Contain(filePath, "goravel\nframework"))
	assert.Equal(t, 2, file.GetLineNum(filePath))
	// Overwrite content
	assert.NoError(t, PutContent(filePath, "welcome", WithMode(0644)))
	assert.False(t, Contain(filePath, "goravel\nframework"))
	assert.True(t, Contain(filePath, "welcome"))
	assert.Equal(t, 1, file.GetLineNum(filePath))
}
