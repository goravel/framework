package http

import (
	"github.com/gin-gonic/gin"
	"github.com/goravel/framework/contracts/http"
)

type Application struct {
}

func (app *Application) Init() (http.Request, http.Response) {
	return request{}, response{}
}

type request struct {
}

func (r request) Validate(c *gin.Context, request http.FormRequest) []error {
	if err := c.ShouldBind(request); err != nil {
	}

	return nil
}

type response struct {
}

func (r response) Success(c *gin.Context, data interface{}) {
	r.Custom(c, data, 200)
}

func (r response) Custom(c *gin.Context, data interface{}, code int) {
	c.JSON(code, data)
}
