package support

import "io"

type Body interface {
	SetFields(fields map[string]any) Body
	SetField(key string, value any) Body
	GetField(key string) any
	SetFile(fieldName, filePath string) Body
	Build() (io.Reader, error)
}
