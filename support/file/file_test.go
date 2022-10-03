package file

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/testing/file"
)

func TestCreate(t *testing.T) {
	pwd, _ := os.Getwd()
	path := pwd + "/goravel/goravel.txt"
	Create(path, `goravel`)
	assert.Equal(t, 1, file.GetLineNum(path))
	assert.True(t, Exist(path))
	assert.True(t, Remove(path))
	assert.True(t, Remove(pwd+"/goravel"))
}
