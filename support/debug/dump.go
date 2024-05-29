package debug

import (
	"io"

	"github.com/davecgh/go-spew/spew"
)

// Dump is used to display detailed information about variables
// And this is a wrapper around spew.Dump.
func Dump(v ...any) {
	spew.Dump(v...)
}

// FDump is used to display detailed information about variables to the specified io.Writer
// And this is a wrapper around spew.Fdump.
func FDump(w io.Writer, v ...any) {
	spew.Fdump(w, v...)
}

// SDump is used to display detailed information about variables as a string,
// And this is a wrapper around spew.Sdump.
func SDump(v ...any) string {
	return spew.Sdump(v...)
}
