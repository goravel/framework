package testing

import (
	"os"
	"strings"
)

func RunInTest() bool {
	return strings.HasSuffix(os.Args[0], ".test")
}
