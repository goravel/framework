package http

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	httpcontract "github.com/goravel/framework/contracts/http"
)

type GinRequest struct {
	instance *gin.Context
	ctx      context.Context
}

func NewGinRequest(instance *gin.Context) *GinRequest {
	return &GinRequest{instance: instance}
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

func (r *GinRequest) File(name string) (httpcontract.File, error) {
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

func (r *GinRequest) WithContext(ctx context.Context) httpcontract.Request {
	r.ctx = ctx

	return r
}

func (r *GinRequest) Method() string {
	return r.instance.Request.Method
}

func (r *GinRequest) Context() context.Context {
	if r.ctx == nil {
		return context.Background()
	}

	return r.ctx
}

func (r *GinRequest) AbortWithStatus(code int) {
	r.instance.AbortWithStatus(code)
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
