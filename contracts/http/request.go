package http

import "github.com/gin-gonic/gin"

type Request interface {
	Validate(c *gin.Context, request FormRequest) []error
}
