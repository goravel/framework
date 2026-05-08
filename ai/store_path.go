package ai

import "strings"

func hasParentPathSegment(path string) bool {
	for _, segment := range strings.Split(path, "/") {
		if segment == ".." {
			return true
		}
	}

	return false
}
