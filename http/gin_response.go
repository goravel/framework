package http

import (
	"bytes"
	"net/http"

	"github.com/gin-gonic/gin"

	httpcontract "github.com/goravel/framework/contracts/http"
)

type GinResponse struct {
	instance *gin.Context
	origin   httpcontract.ResponseOrigin
}

func NewGinResponse(instance *gin.Context, origin httpcontract.ResponseOrigin) *GinResponse {
	return &GinResponse{instance, origin}
}

func (r *GinResponse) String(code int, format string, values ...interface{}) {
	r.instance.String(code, format, values...)
}

func (r *GinResponse) Json(code int, obj interface{}) {
	r.instance.JSON(code, obj)
}

func (r *GinResponse) File(filepath string) {
	r.instance.File(filepath)
}

func (r *GinResponse) Download(filepath, filename string) {
	r.instance.FileAttachment(filepath, filename)
}

func (r *GinResponse) Success() httpcontract.ResponseSuccess {
	return NewGinSuccess(r.instance)
}

func (r *GinResponse) Header(key, value string) httpcontract.Response {
	r.instance.Header(key, value)

	return r
}

func (r *GinResponse) Origin() httpcontract.ResponseOrigin {
	return r.origin
}

type GinSuccess struct {
	instance *gin.Context
}

func NewGinSuccess(instance *gin.Context) httpcontract.ResponseSuccess {
	return &GinSuccess{instance}
}

func (r *GinSuccess) String(format string, values ...interface{}) {
	r.instance.String(http.StatusOK, format, values...)
}

func (r *GinSuccess) Json(obj interface{}) {
	r.instance.JSON(http.StatusOK, obj)
}

func GinResponseMiddleware() httpcontract.Middleware {
	return func(ctx httpcontract.Context) {
		blw := &BodyWriter{body: bytes.NewBufferString("")}
		switch ctx.(type) {
		case *GinContext:
			blw.ResponseWriter = ctx.(*GinContext).Instance().Writer
			ctx.(*GinContext).Instance().Writer = blw
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
