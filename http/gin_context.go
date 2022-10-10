package http

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/goravel/framework/contracts/http"
)

type GinContext struct {
	instance *gin.Context
}

func NewGinContext(ctx *gin.Context) http.Context {
	return &GinContext{ctx}
}

func (c *GinContext) Request() http.Request {
	return NewGinRequest(c.instance)
}

func (c *GinContext) Response() http.Response {
	return NewGinResponse(c.instance)
}

func (c *GinContext) WithValue(key string, value interface{}) {
	c.instance.Set(key, value)
}

func (c *GinContext) Deadline() (deadline time.Time, ok bool) {
	return c.instance.Deadline()
}

func (c *GinContext) Done() <-chan struct{} {
	return c.instance.Done()
}

func (c *GinContext) Err() error {
	return c.instance.Err()
}

func (c *GinContext) Value(key interface{}) interface{} {
	return c.instance.Value(key)
}
