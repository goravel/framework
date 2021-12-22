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
	assert.Equal(t, 17, GetLineNum(".env"))

	err = os.Remove(".env")
	assert.Nil(t, err)
}
