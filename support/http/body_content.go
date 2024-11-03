package http

import (
	"io"

	"github.com/goravel/framework/contracts/support"
)

type bodyContentImpl struct {
	body        io.Reader
	contentType string
}

func newBodyContent(body io.Reader, contentType string) support.BodyContent {
	return &bodyContentImpl{
		body:        body,
		contentType: contentType,
	}
}

func (r *bodyContentImpl) Reader() io.Reader {
	return r.body
}

func (r *bodyContentImpl) ContentType() string {
	return r.contentType
}
