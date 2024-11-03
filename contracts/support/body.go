package support

import "io"

type Body interface {
	SetFields(fields map[string]any) Body
	SetField(key string, value any) Body
	GetField(key string) any
	SetFiles(files map[string]string) Body
	SetFile(fieldName, filePath string) Body
	Build() (BodyContent, error)
}

type BodyContent interface {
	Reader() io.Reader
	ContentType() string
}
