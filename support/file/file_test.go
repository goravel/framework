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
	assert.Nil(t, err)
	assert.Equal(t, "txt", extension)
}

func TestClientOriginalExtension(t *testing.T) {
	assert.Equal(t, ClientOriginalExtension("logo.png"), "png")
}
