package http

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/gofiber/schema"
	"github.com/gookit/validate"
	"github.com/spf13/cast"

	contractsfilesystem "github.com/goravel/framework/contracts/filesystem"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/log"
	contractsession "github.com/goravel/framework/contracts/session"
	contractsvalidate "github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/filesystem"
	"github.com/goravel/framework/foundation/json"
	"github.com/goravel/framework/validation"
)

type sessionKeyType string

const sessionKey sessionKeyType = "session"

var contextRequestPool = sync.Pool{New: func() any {
	return &ContextRequest{
		log:        LogFacade,
		validation: ValidationFacade,
	}
}}

type ContextRequest struct {
	ctx        *Context
	r          *http.Request
	httpBody   map[string]any
	log        log.Log
	validation contractsvalidate.Validation
}

// NewContextRequest creates a new ContextRequest
func NewContextRequest(ctx *Context, log log.Log, validation contractsvalidate.Validation) contractshttp.ContextRequest {
	request := contextRequestPool.Get().(*ContextRequest)
	httpBody, err := getHttpBody(ctx)
	if err != nil {
		log.Error(fmt.Sprintf("%+v", err))
	}
	request.ctx = ctx
	request.r = ctx.r
	request.httpBody = httpBody
	request.log = log
	request.validation = validation
	return request
}

// Abort aborts the request with a status code
func (r *ContextRequest) Abort(code ...int) {
	realCode := http.StatusInternalServerError
	if len(code) > 0 {
		realCode = code[0]
	}

	r.ctx.w.WriteHeader(realCode)
}

// All returns all the input data available
func (r *ContextRequest) All() map[string]any {
	var (
		dataMap  = make(map[string]any)
		queryMap = make(map[string]any)
	)

	for key, query := range r.r.URL.Query() {
		queryMap[key] = strings.Join(query, ",")
	}

	for k, v := range queryMap {
		dataMap[k] = v
	}
	for k, v := range r.httpBody {
		dataMap[k] = v
	}

	return dataMap
}

// Bind binds request data to the provided struct
func (r *ContextRequest) Bind(obj any) error {
	if r.r.Body == nil {
		return errors.New("request body is empty")
	}

	contentType := r.r.Header.Get("Content-Type")

	if strings.Contains(contentType, "application/json") {
		bodyBytes, err := io.ReadAll(r.r.Body)
		if err != nil {
			return err
		}

		// Reset the body for future reads
		r.r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		if len(bodyBytes) == 0 {
			return nil
		}

		return json.NewJson().Unmarshal(bodyBytes, obj)
	} else if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		if err := r.r.ParseForm(); err != nil {
			return err
		}
		decoder := schema.NewDecoder()
		return decoder.Decode(obj, r.r.Form)
	} else if strings.Contains(contentType, "multipart/form-data") {
		if err := r.r.ParseMultipartForm(32 << 20); err != nil {
			return err
		}
		decoder := schema.NewDecoder()
		return decoder.Decode(obj, r.r.Form)
	}

	return errors.New("unsupported media type")
}

// BindQuery binds query parameters to the provided struct
func (r *ContextRequest) BindQuery(obj any) error {
	decoder := schema.NewDecoder()
	return decoder.Decode(obj, r.r.URL.Query())
}

// Cookie retrieves a cookie by key
func (r *ContextRequest) Cookie(key string, defaultValue ...string) string {
	cookie, err := r.r.Cookie(key)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return ""
	}
	return cookie.Value
}

// Form retrieves a form value by key
func (r *ContextRequest) Form(key string, defaultValue ...string) string {
	if err := r.r.ParseForm(); err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return ""
	}

	if value := r.r.PostFormValue(key); value != "" {
		return value
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return ""
}

// File retrieves a file from the request
func (r *ContextRequest) File(name string) (contractsfilesystem.File, error) {
	if err := r.r.ParseMultipartForm(32 << 20); err != nil {
		return nil, err
	}

	_, header, err := r.r.FormFile(name)
	if err != nil {
		return nil, err
	}

	return filesystem.NewFileFromRequest(header)
}

// Files retrieves multiple files from the request
func (r *ContextRequest) Files(name string) ([]contractsfilesystem.File, error) {
	if err := r.r.ParseMultipartForm(32 << 20); err != nil {
		return nil, err
	}

	if r.r.MultipartForm == nil || r.r.MultipartForm.File == nil {
		return nil, http.ErrMissingFile
	}

	if files, ok := r.r.MultipartForm.File[name]; ok && len(files) > 0 {
		var result []contractsfilesystem.File
		for i := range files {
			file, err := filesystem.NewFileFromRequest(files[i])
			if err != nil {
				return nil, err
			}
			result = append(result, file)
		}

		return result, nil
	}

	return nil, http.ErrMissingFile
}

// FullUrl returns the full request URL
func (r *ContextRequest) FullUrl() string {
	prefix := "https://"
	if r.r.TLS == nil {
		prefix = "http://"
	}

	if r.r.Host == "" {
		return ""
	}

	return prefix + r.r.Host + r.r.RequestURI
}

// Header retrieves a header value by key
func (r *ContextRequest) Header(key string, defaultValue ...string) string {
	header := r.r.Header.Get(key)
	if header != "" {
		return header
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return ""
}

// Headers returns all request headers
func (r *ContextRequest) Headers() http.Header {
	return r.r.Header
}

// Host returns the request host
func (r *ContextRequest) Host() string {
	return r.r.Host
}

// HasSession checks if there's a session in the context
func (r *ContextRequest) HasSession() bool {
	_, ok := r.ctx.Ctx.Value(sessionKey).(contractsession.Session)
	return ok
}

// Json retrieves a JSON value by key
func (r *ContextRequest) Json(key string, defaultValue ...string) string {
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

// Method returns the request method
func (r *ContextRequest) Method() string {
	return r.r.Method
}

// Query retrieves a query parameter by key
func (r *ContextRequest) Query(key string, defaultValue ...string) string {
	if values, ok := r.r.URL.Query()[key]; ok && len(values) > 0 {
		return values[0]
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return ""
}

// QueryInt retrieves a query parameter as an integer
func (r *ContextRequest) QueryInt(key string, defaultValue ...int) int {
	if values, ok := r.r.URL.Query()[key]; ok && len(values) > 0 {
		return cast.ToInt(values[0])
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return 0
}

// QueryInt64 retrieves a query parameter as an int64
func (r *ContextRequest) QueryInt64(key string, defaultValue ...int64) int64 {
	if values, ok := r.r.URL.Query()[key]; ok && len(values) > 0 {
		return cast.ToInt64(values[0])
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return 0
}

// QueryBool retrieves a query parameter as a boolean
func (r *ContextRequest) QueryBool(key string, defaultValue ...bool) bool {
	if values, ok := r.r.URL.Query()[key]; ok && len(values) > 0 {
		return stringToBool(values[0])
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return false
}

// QueryArray retrieves a query parameter as an array
func (r *ContextRequest) QueryArray(key string) []string {
	return r.r.URL.Query()[key]
}

// QueryMap retrieves a query parameter as a map
func (r *ContextRequest) QueryMap(key string) map[string]string {
	result := make(map[string]string)
	prefix := key + "["
	suffix := "]"

	for paramKey, paramValues := range r.r.URL.Query() {
		if strings.HasPrefix(paramKey, prefix) && strings.HasSuffix(paramKey, suffix) {
			mapKey := strings.TrimPrefix(paramKey, prefix)
			mapKey = strings.TrimSuffix(mapKey, suffix)
			if len(paramValues) > 0 {
				result[mapKey] = paramValues[0]
			}
		}
	}

	return result
}

// Queries returns all query parameters
func (r *ContextRequest) Queries() map[string]string {
	queries := make(map[string]string)

	for key, query := range r.r.URL.Query() {
		queries[key] = strings.Join(query, ",")
	}

	return queries
}

// Origin returns the original http.Request
func (r *ContextRequest) Origin() *http.Request {
	return r.r
}

// Path returns the request path
func (r *ContextRequest) Path() string {
	return r.r.URL.Path
}

// Input retrieves input from various sources (query, body)
func (r *ContextRequest) Input(key string, defaultValue ...string) string {
	valueFromHttpBody := r.getValueFromHttpBody(key)
	if valueFromHttpBody != nil {
		switch reflect.ValueOf(valueFromHttpBody).Kind() {
		case reflect.Map:
			valueFromHttpBodyObByte, err := json.NewJson().Marshal(valueFromHttpBody)
			if err != nil {
				return ""
			}
			return string(valueFromHttpBodyObByte)
		case reflect.Slice:
			return strings.Join(cast.ToStringSlice(valueFromHttpBody), ",")
		default:
			return cast.ToString(valueFromHttpBody)
		}
	}

	if values, ok := r.r.URL.Query()[key]; ok && len(values) > 0 {
		return values[0]
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return ""
}

// InputArray retrieves input as an array
func (r *ContextRequest) InputArray(key string, defaultValue ...[]string) []string {
	if valueFromHttpBody := r.getValueFromHttpBody(key); valueFromHttpBody != nil {
		if value := cast.ToStringSlice(valueFromHttpBody); value == nil {
			return []string{}
		} else {
			return value
		}
	}

	if values := r.r.URL.Query()[key]; len(values) > 0 {
		if len(values) == 1 && values[0] == "" {
			return []string{}
		}
		return values
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return []string{}
}

// InputMap retrieves input as a map
func (r *ContextRequest) InputMap(key string, defaultValue ...map[string]any) map[string]any {
	if valueFromHttpBody := r.getValueFromHttpBody(key); valueFromHttpBody != nil {
		return cast.ToStringMap(valueFromHttpBody)
	}

	if _, ok := r.r.URL.Query()[key]; ok {
		queryMap := r.QueryMap(key)
		result := make(map[string]any)
		for k, v := range queryMap {
			result[k] = v
		}
		return result
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return map[string]any{}
}

// InputMapArray retrieves input as an array of maps
func (r *ContextRequest) InputMapArray(key string, defaultValue ...[]map[string]any) []map[string]any {
	if valueFromHttpBody := r.getValueFromHttpBody(key); valueFromHttpBody != nil {
		var result = make([]map[string]any, 0)
		for _, item := range cast.ToSlice(valueFromHttpBody) {
			res, err := cast.ToStringMapE(item)
			if err != nil {
				return []map[string]any{}
			}
			result = append(result, res)
		}

		if len(result) == 0 {
			for _, item := range cast.ToStringSlice(valueFromHttpBody) {
				res, err := cast.ToStringMapE(item)
				if err != nil {
					return []map[string]any{}
				}
				result = append(result, res)
			}
		}

		return result
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return []map[string]any{}
}

// InputInt retrieves input as an integer
func (r *ContextRequest) InputInt(key string, defaultValue ...int) int {
	value := r.Input(key)
	if value == "" && len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return cast.ToInt(value)
}

// InputInt64 retrieves input as an int64
func (r *ContextRequest) InputInt64(key string, defaultValue ...int64) int64 {
	value := r.Input(key)
	if value == "" && len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return cast.ToInt64(value)
}

// InputBool retrieves input as a boolean
func (r *ContextRequest) InputBool(key string, defaultValue ...bool) bool {
	value := r.Input(key)
	if value == "" && len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return stringToBool(value)
}

// Ip returns the client IP address
func (r *ContextRequest) Ip() string {
	// Check for X-Forwarded-For header first
	ip := r.r.Header.Get("X-Forwarded-For")
	if ip != "" {
		// The X-Forwarded-For header can contain multiple IPs
		// The leftmost one is the original client IP
		parts := strings.Split(ip, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	// Check for X-Real-IP header
	ip = r.r.Header.Get("X-Real-Ip")
	if ip != "" {
		return ip
	}

	// Fall back to RemoteAddr
	if r.r.RemoteAddr != "" {
		// RemoteAddr is in the format "IP:port"
		ipPort := strings.Split(r.r.RemoteAddr, ":")
		if len(ipPort) > 0 {
			return ipPort[0]
		}
	}

	return ""
}

// Route retrieves a route parameter
func (r *ContextRequest) Route(key string) string {
	return ""
}

// RouteInt retrieves a route parameter as an integer
func (r *ContextRequest) RouteInt(key string) int {
	return 0
}

// RouteInt64 retrieves a route parameter as an int64
func (r *ContextRequest) RouteInt64(key string) int64 {
	return 0
}

// Session returns the session from the context
func (r *ContextRequest) Session() contractsession.Session {
	s, ok := r.ctx.Ctx.Value(sessionKey).(contractsession.Session)
	if !ok {
		return nil
	}
	return s
}

// SetSession sets the session in the context
func (r *ContextRequest) SetSession(session contractsession.Session) contractshttp.ContextRequest {
	r.ctx.Ctx = context.WithValue(r.ctx.Ctx, sessionKey, session)
	return r
}

// Url returns the request URL
func (r *ContextRequest) Url() string {
	return r.r.RequestURI
}

// Validate validates the request data against the provided rules
func (r *ContextRequest) Validate(rules map[string]string, options ...contractsvalidate.Option) (contractsvalidate.Validator, error) {
	if len(rules) == 0 {
		return nil, errors.New("rules can't be empty")
	}

	options = append(options, validation.Rules(rules), validation.CustomRules(r.validation.Rules()), validation.CustomFilters(r.validation.Filters()))

	dataFace, err := validate.FromRequest(r.r)
	if err != nil {
		return nil, err
	}

	for key, query := range r.r.URL.Query() {
		if _, exist := dataFace.Get(key); !exist {
			if _, err := dataFace.Set(key, strings.Join(query, ",")); err != nil {
				return nil, err
			}
		}
	}

	return r.validation.Make(dataFace, rules, options...)
}

// ValidateRequest validates a FormRequest
func (r *ContextRequest) ValidateRequest(request contractshttp.FormRequest) (contractsvalidate.Errors, error) {
	if err := request.Authorize(r.ctx); err != nil {
		return nil, err
	}

	var options []contractsvalidate.Option
	if requestWithFilters, ok := request.(contractshttp.FormRequestWithFilters); ok {
		options = append(options, validation.Filters(requestWithFilters.Filters(r.ctx)))
	}
	if requestWithMessage, ok := request.(contractshttp.FormRequestWithMessages); ok {
		options = append(options, validation.Messages(requestWithMessage.Messages(r.ctx)))
	}
	if requestWithAttributes, ok := request.(contractshttp.FormRequestWithAttributes); ok {
		options = append(options, validation.Attributes(requestWithAttributes.Attributes(r.ctx)))
	}
	if prepareForValidation, ok := request.(contractshttp.FormRequestWithPrepareForValidation); ok {
		options = append(options, validation.PrepareForValidation(r.ctx, prepareForValidation.PrepareForValidation))
	}

	validator, err := r.Validate(request.Rules(r.ctx), options...)
	if err != nil {
		return nil, err
	}

	if err := validator.Bind(request); err != nil {
		return nil, err
	}

	return validator.Errors(), nil
}

// getValueFromHttpBody retrieves a value from the HTTP body
func (r *ContextRequest) getValueFromHttpBody(key string) any {
	if r.httpBody == nil {
		return nil
	}

	var current any
	current = r.httpBody
	keys := strings.Split(key, ".")
	for _, k := range keys {
		currentValue := reflect.ValueOf(current)
		switch currentValue.Kind() {
		case reflect.Map:
			if value := currentValue.MapIndex(reflect.ValueOf(k)); value.IsValid() {
				current = value.Interface()
			} else {
				if value := currentValue.MapIndex(reflect.ValueOf(k + "[]")); value.IsValid() {
					current = value.Interface()
				} else {
					return nil
				}
			}
		case reflect.Slice:
			if number, err := strconv.Atoi(k); err == nil {
				return cast.ToStringSlice(current)[number]
			} else {
				return nil
			}
		}
	}

	return current
}

// getHttpBody parses the request body
func getHttpBody(ctx *Context) (map[string]any, error) {
	request := ctx.r
	if request == nil || request.Body == nil || request.ContentLength == 0 {
		return nil, nil
	}

	contentType := request.Header.Get("Content-Type")
	data := make(map[string]any)

	if strings.Contains(contentType, "application/json") {
		bodyBytes, err := io.ReadAll(request.Body)
		_ = request.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("retrieve json error: %v", err)
		}

		if len(bodyBytes) > 0 {
			if err = json.NewJson().Unmarshal(bodyBytes, &data); err != nil {
				return nil, fmt.Errorf("decode json [%v] error: %v", string(bodyBytes), err)
			}

			request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
	}

	if strings.Contains(contentType, "multipart/form-data") {
		if err := request.ParseMultipartForm(32 << 20); err != nil {
			return nil, fmt.Errorf("parse multipart form error: %v", err)
		}

		if request.MultipartForm != nil {
			for k, v := range request.MultipartForm.Value {
				if len(v) > 1 {
					data[k] = v
				} else if len(v) == 1 {
					data[k] = v[0]
				}
			}

			if request.MultipartForm.File != nil {
				for k, v := range request.MultipartForm.File {
					if len(v) > 1 {
						data[k] = v
					} else if len(v) == 1 {
						data[k] = v[0]
					}
				}
			}
		}
	}

	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		if err := request.ParseForm(); err != nil {
			return nil, fmt.Errorf("parse form error: %v", err)
		}

		for k, v := range request.Form {
			if len(v) > 1 {
				data[k] = v
			} else if len(v) == 1 {
				data[k] = v[0]
			}
		}
	}

	return data, nil
}

// stringToBool converts a string to a boolean
func stringToBool(value string) bool {
	return value == "1" || value == "true" || value == "on" || value == "yes"
}
