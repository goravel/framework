package http

import (
	"io"

	"github.com/goravel/framework/contracts/support/http"
)

type readerImpl struct {
	body        io.Reader
	contentType string
}

func newReader(body io.Reader, contentType string) http.Reader {
	return &readerImpl{
		body:        body,
		contentType: contentType,
	}
}

func (r *readerImpl) Reader() io.Reader {
	return r.body
}

func (r *readerImpl) ContentType() string {
	return r.contentType
}
