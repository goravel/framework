package debug

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func redirectStdout(fn func()) ([]byte, error) {
	f, err := os.CreateTemp("", "stdout")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f.Name())
	defer f.Close()

	orig := os.Stdout
	os.Stdout = f
	fn()
	os.Stdout = orig

	return os.ReadFile(f.Name())
}

func TestDump(t *testing.T) {
	buf, err := redirectStdout(func() {
		Dump("foo")
	})
	assert.NoError(t, err)
	assert.Equal(t, `(string) (len=3) "foo"
`, string(buf))
}

func TestFDump(t *testing.T) {
	var buf bytes.Buffer
	w := &buf

	FDump(w, "foo")

	assert.Equal(t, `(string) (len=3) "foo"
`, buf.String())
}

func TestSDump(t *testing.T) {
	assert.Equal(t, `(string) (len=3) "foo"
`, SDump("foo"))
}
