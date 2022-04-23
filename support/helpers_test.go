package support

import (
	testing2 "github.com/goravel/framework/support/testing"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetLineNum(t *testing.T) {
	err := testing2.CreateEnv()
	assert.Nil(t, err)
	assert.Equal(t, 18, Helpers{}.GetLineNum(".env"))

	err = os.Remove(".env")
	assert.Nil(t, err)
}

func TestCreateFile(t *testing.T) {
	pwd, _ := os.Getwd()
	path := pwd + "/goravel/goravel.txt"
	Helpers{}.CreateFile(path, `goravel`)
	assert.Equal(t, 1, Helpers{}.GetLineNum(path))
	assert.True(t, Helpers{}.ExistFile(path))
	assert.True(t, Helpers{}.RemoveFile(path))
	assert.True(t, Helpers{}.RemoveFile(pwd+"/goravel"))
}

func TestCase2Camel(t *testing.T) {
	assert.Equal(t, "GoravelFramework", Helpers{}.Case2Camel("goravel_framework"))
	assert.Equal(t, "GoravelFramework1", Helpers{}.Case2Camel("goravel_framework1"))
}

func TestCamel2Case(t *testing.T) {
	assert.Equal(t, "goravel_framework", Helpers{}.Camel2Case("GoravelFramework"))
	assert.Equal(t, "goravel_framework1", Helpers{}.Camel2Case("GoravelFramework1"))
}
