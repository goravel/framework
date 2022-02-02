package http

import "github.com/gin-gonic/gin"

type Response interface {
	Success(c *gin.Context, data interface{})
	Custom(c *gin.Context, data interface{}, code int)
}
