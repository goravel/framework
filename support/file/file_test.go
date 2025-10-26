package file

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/testing/file"
)

func TestClientOriginalExtension(t *testing.T) {
	assert.Equal(t, ClientOriginalExtension("logo.png"), "png")
}

func TestContain(t *testing.T) {
	assert.True(t, Contain("../constant.go", "Version"))
}

func TestCreate(t *testing.T) {
	filePath := path.Join(t.TempDir(), "goravel.txt")
	assert.Nil(t, PutContent(filePath, `goravel`))
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
	assert.Nil(t, PutContent(filePath, `goravel`))

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
	assert.Contains(t, content, "Version")
}

func TestPutContent(t *testing.T) {
	if !env.IsWindows() {
		// directory creation failure
		assert.Error(t, PutContent("/proc/invalid/file.txt", "content"))
		// write failure (create read-only dir)
		readOnlyDir := path.Join(t.TempDir(), "readonly")
		assert.NoError(t, os.Mkdir(readOnlyDir, 0444))
		assert.Error(t, PutContent(path.Join(readOnlyDir, "file.txt"), "content"))
	}
	// create a file and put content
	filePath := path.Join(t.TempDir(), "goravel.txt")
	assert.NoError(t, PutContent(filePath, "goravel"))
	assert.True(t, Contain(filePath, "goravel"))
	assert.Equal(t, 1, file.GetLineNum(filePath))
	// append content
	assert.NoError(t, PutContent(filePath, "\nframework", WithAppend()))
	assert.True(t, Contain(filePath, "goravel\nframework"))
	assert.Equal(t, 2, file.GetLineNum(filePath))
	// overwrite content
	assert.NoError(t, PutContent(filePath, "welcome", WithMode(0644)))
	assert.False(t, Contain(filePath, "goravel\nframework"))
	assert.True(t, Contain(filePath, "welcome"))
	assert.Equal(t, 1, file.GetLineNum(filePath))
}
