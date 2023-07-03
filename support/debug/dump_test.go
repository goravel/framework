package debug

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestDump(t *testing.T) {
	Dump("foo")
}

func TestFDump(t *testing.T) {
	FDump(os.Stdout, "foo")

	fmt.Sprintf("%s", SDump("foo"))
}

func TestSDump(t *testing.T) {
	assert.Equal(t, `(string) (len=3) "foo"
`, SDump("foo"))
}
