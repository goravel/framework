package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	httpcontract "github.com/goravel/framework/contracts/http"
)

type GinResponse struct {
	instance *gin.Context
}

func NewGinResponse(instance *gin.Context) httpcontract.Response {
	return &GinResponse{instance: instance}
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
