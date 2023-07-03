package debug

import (
	"github.com/davecgh/go-spew/spew"
	"io"
)

// Dump is a wrapper around spew.Dump.
func Dump(v ...interface{}) {
	spew.Dump(v...)
}

// FDump is a wrapper around spew.Fdump.
func FDump(w io.Writer, v ...interface{}) {
	spew.Fdump(w, v...)
}

// SDump is a wrapper around spew.Sdump.
func SDump(v ...interface{}) string {
	return spew.Sdump(v...)
}
