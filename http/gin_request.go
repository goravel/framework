package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	contracthttp "github.com/goravel/framework/contracts/http"
)

type GinRequest struct {
	instance *gin.Context
}

func NewGinRequest(instance *gin.Context) contracthttp.Request {
	return &GinRequest{instance}
}

func (r *GinRequest) Input(key string) string {
	return r.instance.Param(key)
}

func (r *GinRequest) Query(key, defaultValue string) string {
	return r.instance.DefaultQuery(key, defaultValue)
}

func (r *GinRequest) Form(key, defaultValue string) string {
	return r.instance.DefaultPostForm(key, defaultValue)
}

func (r *GinRequest) Bind(obj interface{}) error {
	return r.instance.ShouldBind(obj)
}

func (r *GinRequest) File(name string) (contracthttp.File, error) {
	file, err := r.instance.FormFile(name)
	if err != nil {
		return nil, err
	}

	return &GinFile{instance: r.instance, file: file}, nil
}

func (r *GinRequest) Header(key, defaultValue string) string {
	header := r.instance.GetHeader(key)
	if header != "" {
		return header
	}

	return defaultValue
}

func (r *GinRequest) Headers() http.Header {
	return r.instance.Request.Header
}

func (r *GinRequest) Method() string {
	return r.instance.Request.Method
}

func (r *GinRequest) Url() string {
	return r.instance.Request.RequestURI
}

func (r *GinRequest) FullUrl() string {
	prefix := "https://"
	if r.instance.Request.TLS == nil {
		prefix = "http://"
	}

	if r.instance.Request.Host == "" {
		return ""
	}

	return prefix + r.instance.Request.Host + r.instance.Request.RequestURI
}

func (r *GinRequest) AbortWithStatus(code int) {
	r.instance.AbortWithStatus(code)
}

func (r *GinRequest) AbortWithStatusJSON(code int, jsonObj interface{}) {
	r.instance.AbortWithStatusJSON(code, jsonObj)
}

func (r *GinRequest) Next() {
	r.instance.Next()
}

func (r *GinRequest) Path() string {
	return r.instance.Request.URL.Path
}

func (r *GinRequest) Ip() string {
	return r.instance.ClientIP()
}

func (r *GinRequest) Origin() *http.Request {
	return r.instance.Request
}

func (r *GinRequest) Response() contracthttp.Response {
	return NewGinResponse(r.instance)
}
