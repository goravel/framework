package http

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	httpcontract "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

type GinRequest struct {
	ctx      context.Context
	instance *gin.Context
}

func NewGinRequest(instance *gin.Context) httpcontract.Request {
	if facades.Request == nil {
		facades.Request = &GinRequest{ctx: context.Background(), instance: instance}
	} else {
		facades.Request = &GinRequest{ctx: facades.Request.Context(), instance: instance}
	}

	return facades.Request
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

func (r *GinRequest) Url() string {
	return r.instance.Request.RequestURI
}

func (r *GinRequest) FullUrl() string {
	http := "https://"
	if r.instance.Request.TLS == nil {
		http = "http://"
	}

	return http + r.instance.Request.Host + r.instance.Request.RequestURI
}

func (r *GinRequest) Context() context.Context {
	if r.ctx == nil {
		r.ctx = context.Background()
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
