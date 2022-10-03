package file

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateEnv(t *testing.T) {
	err := CreateEnv()
	assert.Nil(t, err)
	assert.FileExists(t, ".env")

	err = os.Remove(".env")
	assert.Nil(t, err)
}

func TestGetLineNum(t *testing.T) {
	err := CreateEnv()
	assert.Nil(t, err)
	assert.Equal(t, 18, GetLineNum(".env"))

	err = os.Remove(".env")
	assert.Nil(t, err)
}
