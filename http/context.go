package http

import (
	"github.com/gin-gonic/gin"
	"github.com/goravel/framework/contracts/http"
)

func Background() http.Context {
	return NewGinContext(&gin.Context{})
}
