package http

import (
	"mime/multipart"

	"github.com/gin-gonic/gin"
)

type GinFile struct {
	instance *gin.Context
	file     *multipart.FileHeader
}

func (f *GinFile) Store(dst string) error {
	return f.instance.SaveUploadedFile(f.file, dst)
}
