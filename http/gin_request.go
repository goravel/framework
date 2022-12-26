package http

import (
	"errors"
	"net/http"

	contractsfilesystem "github.com/goravel/framework/contracts/filesystem"
	httpcontract "github.com/goravel/framework/contracts/http"
	validatecontract "github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/filesystem"
	"github.com/goravel/framework/validation"

	"github.com/gin-gonic/gin"
	"github.com/gookit/validate"
)

type GinRequest struct {
	ctx      *GinContext
	instance *gin.Context
}

func NewGinRequest(ctx *GinContext) httpcontract.Request {
	return &GinRequest{ctx, ctx.instance}
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

func (r *GinRequest) Bind(obj any) error {
	return r.instance.ShouldBind(obj)
}

func (r *GinRequest) File(name string) (contractsfilesystem.File, error) {
	file, err := r.instance.FormFile(name)
	if err != nil {
		return nil, err
	}

	return filesystem.NewFileFromRequest(file)
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

func (r *GinRequest) Method() string {
	return r.instance.Request.Method
}

func (r *GinRequest) Url() string {
	return r.instance.Request.RequestURI
}

func (r *GinRequest) FullUrl() string {
	prefix := "https://"
	if r.instance.Request.TLS == nil {
		prefix = "http://"
	}

	if r.instance.Request.Host == "" {
		return ""
	}

	return prefix + r.instance.Request.Host + r.instance.Request.RequestURI
}

func (r *GinRequest) AbortWithStatus(code int) {
	r.instance.AbortWithStatus(code)
}

func (r *GinRequest) AbortWithStatusJson(code int, jsonObj any) {
	r.instance.AbortWithStatusJSON(code, jsonObj)
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

func (r *GinRequest) Origin() *http.Request {
	return r.instance.Request
}

func (r *GinRequest) Response() httpcontract.Response {
	return NewGinResponse(r.instance)
}

func (r *GinRequest) Validate(rules map[string]string, options ...validatecontract.Option) (validatecontract.Validator, error) {
	if rules == nil || len(rules) == 0 {
		return nil, errors.New("rules can't be empty")
	}

	options = append(options, validation.Rules(rules), validation.CustomRules(facades.Validation.Rules()))
	generateOptions := validation.GenerateOptions(options)

	var v *validate.Validation
	dataFace, err := validate.FromRequest(r.Origin())
	if err != nil {
		return nil, err
	}
	if dataFace == nil {
		v = validate.NewValidation(dataFace)
	} else {
		if generateOptions["prepareForValidation"] != nil {
			generateOptions["prepareForValidation"].(func(data validatecontract.Data))(validation.NewData(dataFace))
		}

		v = dataFace.Create()
	}

	validation.AppendOptions(v, generateOptions)

	return validation.NewValidator(v, dataFace), nil
}

func (r *GinRequest) ValidateRequest(request httpcontract.FormRequest) (validatecontract.Errors, error) {
	if err := request.Authorize(r.ctx); err != nil {
		return nil, err
	}

	validator, err := r.Validate(request.Rules(), validation.Messages(request.Messages()), validation.Attributes(request.Attributes()), validation.PrepareForValidation(request.PrepareForValidation))
	if err != nil {
		return nil, err
	}

	if err := validator.Bind(request); err != nil {
		return nil, err
	}

	return validator.Errors(), nil
}
