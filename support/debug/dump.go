package debug

import (
	"io"

	"github.com/davecgh/go-spew/spew"
)

// Dump is used to display detailed information about variables
// And this is a wrapper around spew.Dump.
func Dump(v ...interface{}) {
	spew.Dump(v...)
}

// FDump is used to display detailed information about variables to the specified io.Writer
// And this is a wrapper around spew.Fdump.
func FDump(w io.Writer, v ...interface{}) {
	spew.Fdump(w, v...)
}

// SDump is used to display detailed information about variables as a string,
// And this is a wrapper around spew.Sdump.
func SDump(v ...interface{}) string {
	return spew.Sdump(v...)
}
