package http

import "github.com/gin-gonic/gin"

type Response interface {
	Success(ctx *gin.Context, data interface{})
	Custom(ctx *gin.Context, data interface{}, code int)
}
