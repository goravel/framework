package http

import "github.com/gin-gonic/gin"

type Request interface {
	Validate(ctx *gin.Context, request FormRequest) []error
}
