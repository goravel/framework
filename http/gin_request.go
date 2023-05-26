package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gookit/validate"
	"github.com/spf13/cast"

	filesystemcontract "github.com/goravel/framework/contracts/filesystem"
	httpcontract "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/log"
	validatecontract "github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/filesystem"
	"github.com/goravel/framework/validation"
)

type GinRequest struct {
	ctx        *GinContext
	instance   *gin.Context
	postData   map[string]any
	log        log.Log
	validation validatecontract.Validation
}

func NewGinRequest(ctx *GinContext, log log.Log, validation validatecontract.Validation) httpcontract.Request {
	postData, err := getPostData(ctx)
	if err != nil {
		LogFacade.Error(fmt.Sprintf("%+v", errors.Unwrap(err)))
	}

	return &GinRequest{ctx: ctx, instance: ctx.instance, postData: postData, log: log, validation: validation}
}

func (r *GinRequest) AbortWithStatus(code int) {
	r.instance.AbortWithStatus(code)
}

func (r *GinRequest) AbortWithStatusJson(code int, jsonObj any) {
	r.instance.AbortWithStatusJSON(code, jsonObj)
}

func (r *GinRequest) All() map[string]any {
	var (
		dataMap  = make(map[string]any)
		queryMap = make(map[string]any)
	)

	for key, query := range r.instance.Request.URL.Query() {
		queryMap[key] = strings.Join(query, ",")
	}

	var mu sync.RWMutex
	for k, v := range queryMap {
		mu.Lock()
		dataMap[k] = v
		mu.Unlock()
	}
	for k, v := range r.postData {
		mu.Lock()
		dataMap[k] = v
		mu.Unlock()
	}

	return dataMap
}

func (r *GinRequest) Bind(obj any) error {
	return r.instance.ShouldBind(obj)
}

func (r *GinRequest) Form(key string, defaultValue ...string) string {
	if len(defaultValue) == 0 {
		return r.instance.PostForm(key)
	}

	return r.instance.DefaultPostForm(key, defaultValue[0])
}

func (r *GinRequest) File(name string) (filesystemcontract.File, error) {
	file, err := r.instance.FormFile(name)
	if err != nil {
		return nil, err
	}

	return filesystem.NewFileFromRequest(file)
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

func (r *GinRequest) Header(key string, defaultValue ...string) string {
	header := r.instance.GetHeader(key)
	if header != "" {
		return header
	}

	if len(defaultValue) == 0 {
		return ""
	}

	return defaultValue[0]
}

func (r *GinRequest) Headers() http.Header {
	return r.instance.Request.Header
}

func (r *GinRequest) Host() string {
	return r.instance.Request.Host
}

func (r *GinRequest) Json(key string, defaultValue ...string) string {
	var data map[string]any
	if err := r.Bind(&data); err != nil {
		if len(defaultValue) == 0 {
			return ""
		} else {
			return defaultValue[0]
		}
	}

	if value, exist := data[key]; exist {
		return cast.ToString(value)
	}

	if len(defaultValue) == 0 {
		return ""
	}

	return defaultValue[0]
}

func (r *GinRequest) Method() string {
	return r.instance.Request.Method
}

func (r *GinRequest) Next() {
	r.instance.Next()
}

func (r *GinRequest) Query(key string, defaultValue ...string) string {
	if len(defaultValue) > 0 {
		return r.instance.DefaultQuery(key, defaultValue[0])
	}

	return r.instance.Query(key)
}

func (r *GinRequest) QueryInt(key string, defaultValue ...int) int {
	if val, ok := r.instance.GetQuery(key); ok {
		return cast.ToInt(val)
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return 0
}

func (r *GinRequest) QueryInt64(key string, defaultValue ...int64) int64 {
	if val, ok := r.instance.GetQuery(key); ok {
		return cast.ToInt64(val)
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return 0
}

func (r *GinRequest) QueryBool(key string, defaultValue ...bool) bool {
	if value, ok := r.instance.GetQuery(key); ok {
		return stringToBool(value)
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return false
}

func (r *GinRequest) QueryArray(key string) []string {
	return r.instance.QueryArray(key)
}

func (r *GinRequest) QueryMap(key string) map[string]string {
	return r.instance.QueryMap(key)
}

func (r *GinRequest) Queries() map[string]string {
	queries := make(map[string]string)

	for key, query := range r.instance.Request.URL.Query() {
		queries[key] = strings.Join(query, ",")
	}

	return queries
}

func (r *GinRequest) Origin() *http.Request {
	return r.instance.Request
}

func (r *GinRequest) Path() string {
	return r.instance.Request.URL.Path
}

func (r *GinRequest) Input(key string, defaultValue ...string) string {
	if value, exist := r.postData[key]; exist {
		return cast.ToString(value)
	}

	if value, exist := r.instance.GetQuery(key); exist {
		return value
	}

	value := r.instance.Param(key)
	if value == "" && len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return value
}

func (r *GinRequest) InputInt(key string, defaultValue ...int) int {
	value := r.Input(key)
	if value == "" && len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return cast.ToInt(value)
}

func (r *GinRequest) InputInt64(key string, defaultValue ...int64) int64 {
	value := r.Input(key)
	if value == "" && len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return cast.ToInt64(value)
}

func (r *GinRequest) InputBool(key string, defaultValue ...bool) bool {
	value := r.Input(key)
	if value == "" && len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return stringToBool(value)
}

func (r *GinRequest) Ip() string {
	return r.instance.ClientIP()
}

func (r *GinRequest) Route(key string) string {
	return r.instance.Param(key)
}

func (r *GinRequest) RouteInt(key string) int {
	val := r.instance.Param(key)

	return cast.ToInt(val)
}

func (r *GinRequest) RouteInt64(key string) int64 {
	val := r.instance.Param(key)

	return cast.ToInt64(val)
}

func (r *GinRequest) Url() string {
	return r.instance.Request.RequestURI
}

func (r *GinRequest) Validate(rules map[string]string, options ...validatecontract.Option) (validatecontract.Validator, error) {
	if len(rules) == 0 {
		return nil, errors.New("rules can't be empty")
	}

	options = append(options, validation.Rules(rules), validation.CustomRules(r.validation.Rules()))
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
			if err := generateOptions["prepareForValidation"].(func(ctx httpcontract.Context, data validatecontract.Data) error)(r.ctx, validation.NewData(dataFace)); err != nil {
				return nil, err
			}
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

	validator, err := r.Validate(request.Rules(r.ctx), validation.Messages(request.Messages(r.ctx)), validation.Attributes(request.Attributes(r.ctx)), func(options map[string]any) {
		options["prepareForValidation"] = request.PrepareForValidation
	})
	if err != nil {
		return nil, err
	}

	if err := validator.Bind(request); err != nil {
		return nil, err
	}

	return validator.Errors(), nil
}

func getPostData(ctx *GinContext) (map[string]any, error) {
	request := ctx.instance.Request
	if request == nil || request.Body == nil || request.ContentLength == 0 {
		return nil, nil
	}

	contentType := ctx.instance.ContentType()
	data := make(map[string]any)
	if contentType == "application/json" {
		bodyBytes, err := ioutil.ReadAll(request.Body)
		_ = request.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("retrieve json error: %v", err)
		}

		if err := json.Unmarshal(bodyBytes, &data); err != nil {
			return nil, fmt.Errorf("decode json [%v] error: %v", string(bodyBytes), err)
		}

		request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	if contentType == "multipart/form-data" {
		if request.PostForm == nil {
			const defaultMemory = 32 << 20
			if err := request.ParseMultipartForm(defaultMemory); err != nil {
				return nil, fmt.Errorf("parse multipart form error: %v", err)
			}
		}
		for k, v := range request.PostForm {
			data[k] = strings.Join(v, ",")
		}
		for k, v := range request.MultipartForm.File {
			if len(v) > 0 {
				data[k] = v[0]
			}
		}
	}

	return data, nil
}

func stringToBool(value string) bool {
	return value == "1" || value == "true" || value == "on" || value == "yes"
}
