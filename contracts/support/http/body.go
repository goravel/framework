package http

import "io"

type Body interface {
	Build() (Reader, error)
	GetField(key string) any
	SetField(key string, value any) Body
	SetFields(fields map[string]any) Body
	SetFile(fieldName, filePath string) Body
	SetFiles(files map[string]string) Body
}

type Reader interface {
	Reader() io.Reader
	ContentType() string
}
