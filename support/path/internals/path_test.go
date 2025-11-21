package internals

import (
	"testing"

	"github.com/goravel/framework/support"
	"github.com/stretchr/testify/assert"
)

func TestPath(t *testing.T) {
	support.RelativePath = "."
	result := Path("foo", "bar.txt")
	expected := Abs(".", "app", "foo", "bar.txt")

	assert.Equal(t, expected, result)
}

func TestFacadesPath(t *testing.T) {
	support.RelativePath = "." // Set to current dir for test
	result := Facades("foo.txt")
	expected := Abs(".", "app", "facades", "foo.txt")

	assert.Equal(t, expected, result)
}
