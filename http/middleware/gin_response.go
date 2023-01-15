package middleware

import (
	"bytes"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/http"

	"github.com/gin-gonic/gin"
)

func GinResponse() contractshttp.Middleware {
	return func(ctx contractshttp.Context) {
		blw := &BodyWriter{body: bytes.NewBufferString("")}
		switch ctx.(type) {
		case *http.GinContext:
			blw.ResponseWriter = ctx.(*http.GinContext).Instance().Writer
			ctx.(*http.GinContext).Instance().Writer = blw
		}

		ctx.WithValue("responseOrigin", blw)
		ctx.Request().Next()
	}
}

type BodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *BodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)

	return w.ResponseWriter.Write(b)
}

func (w *BodyWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)

	return w.ResponseWriter.WriteString(s)
}

func (w *BodyWriter) Body() *bytes.Buffer {
	return w.body
}
