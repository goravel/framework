package support

import "io"

type Body interface {
	SetField(key string, value any) Body
	SetFile(fieldName, filePath string) Body
	Build() (io.Reader, error)
}
