package file

import (
	"os"
	"testing"

	"github.com/goravel/framework/testing/file"

	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	pwd, _ := os.Getwd()
	path := pwd + "/goravel/goravel.txt"
	Create(path, `goravel`)
	assert.Equal(t, 1, file.GetLineNum(path))
	assert.True(t, Exists(path))
	assert.True(t, Remove(path))
	assert.True(t, Remove(pwd+"/goravel"))
}

func TestExtension(t *testing.T) {
	extension, err := Extension("file.go")
	assert.EqualError(t, err, "unknown file extension")
	assert.Empty(t, extension)

	extension, err = Extension("file.go", true)
	assert.Nil(t, err)
	assert.Equal(t, extension, "go")
}

func TestClientOriginalExtension(t *testing.T) {
	assert.Equal(t, ClientOriginalExtension("logo.png"), "png")
}
