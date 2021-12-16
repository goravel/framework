package support

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestRunInTest(t *testing.T) {
	assert.True(t, RunInTest())
}

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
	assert.Equal(t, 13, GetLineNum(".env"))

	err = os.Remove(".env")
	assert.Nil(t, err)
}