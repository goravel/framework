package route

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	mockconfig "github.com/goravel/framework/contracts/config/mocks"
	httpcontract "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/testing/mock"
)

func TestGinRequest(t *testing.T) {
	var (
		gin        *Gin
		req        *http.Request
		mockConfig *mockconfig.Config
	)
	beforeEach := func() {
		mockConfig = mock.Config()
		mockConfig.On("GetBool", "app.debug").Return(true).Once()

		gin = NewGin()
	}
	tests := []struct {
		name       string
		method     string
		url        string
		setup      func(method, url string) error
		expectCode int
		expectBody string
	}{
		{
			name:   "Methods",
			method: "GET",
			url:    "/get/1?name=Goravel",
			setup: func(method, url string) error {
				gin.Get("/get/{id}", func(ctx httpcontract.Context) {
					ctx.Response().Success().Json(httpcontract.Json{
						"id":       ctx.Request().Input("id"),
						"name":     ctx.Request().Query("name", "Hello"),
						"header":   ctx.Request().Header("Hello", "World"),
						"method":   ctx.Request().Method(),
						"path":     ctx.Request().Path(),
						"url":      ctx.Request().Url(),
						"full_url": ctx.Request().FullUrl(),
						"ip":       ctx.Request().Ip(),
					})
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}
				req.Header.Set("Hello", "goravel")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"full_url\":\"\",\"header\":\"goravel\",\"id\":\"1\",\"ip\":\"\",\"method\":\"GET\",\"name\":\"Goravel\",\"path\":\"/get/1\",\"url\":\"\"}",
		},
		{
			name:   "Headers",
			method: "GET",
			url:    "/headers",
			setup: func(method, url string) error {
				gin.Get("/headers", func(ctx httpcontract.Context) {
					str, _ := json.Marshal(ctx.Request().Headers())
					ctx.Response().Success().String(string(str))
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}
				req.Header.Set("Hello", "Goravel")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"Hello\":[\"Goravel\"]}",
		},
		{
			name:   "Form",
			method: "POST",
			url:    "/post",
			setup: func(method, url string) error {
				gin.Post("/post", func(ctx httpcontract.Context) {
					ctx.Response().Success().Json(httpcontract.Json{
						"name": ctx.Request().Form("name", "Hello"),
					})
				})

				payload := &bytes.Buffer{}
				writer := multipart.NewWriter(payload)
				if err := writer.WriteField("name", "Goravel"); err != nil {
					return err
				}
				if err := writer.Close(); err != nil {
					return err
				}

				req, _ = http.NewRequest(method, url, payload)
				req.Header.Set("Content-Type", writer.FormDataContentType())

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"name\":\"Goravel\"}",
		},
		{
			name:   "Bind",
			method: "POST",
			url:    "/bind",
			setup: func(method, url string) error {
				gin.Post("/bind", func(ctx httpcontract.Context) {
					type Test struct {
						Name string
					}
					var test Test
					_ = ctx.Request().Bind(&test)
					ctx.Response().Success().Json(httpcontract.Json{
						"name": test.Name,
					})
				})

				payload := strings.NewReader(`{
					"Name": "Goravel"
				}`)
				req, _ = http.NewRequest(method, url, payload)
				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"name\":\"Goravel\"}",
		},
		{
			name:   "QueryArray",
			method: "GET",
			url:    "/query-array?name=Goravel&name=Goravel1",
			setup: func(method, url string) error {
				gin.Get("/query-array", func(ctx httpcontract.Context) {
					ctx.Response().Success().Json(httpcontract.Json{
						"name": ctx.Request().QueryArray("name"),
					})
				})

				req, _ = http.NewRequest(method, url, nil)
				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"name\":[\"Goravel\",\"Goravel1\"]}",
		},
		{
			name:   "QueryMap",
			method: "GET",
			url:    "/query-array?name[a]=Goravel&name[b]=Goravel1",
			setup: func(method, url string) error {
				gin.Get("/query-array", func(ctx httpcontract.Context) {
					ctx.Response().Success().Json(httpcontract.Json{
						"name": ctx.Request().QueryMap("name"),
					})
				})

				req, _ = http.NewRequest(method, url, nil)
				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"name\":{\"a\":\"Goravel\",\"b\":\"Goravel1\"}}",
		},
		{
			name:   "File",
			method: "POST",
			url:    "/file",
			setup: func(method, url string) error {
				gin.Post("/file", func(ctx httpcontract.Context) {
					mockConfig.On("GetString", "filesystems.default").Return("local").Once()

					fileInfo, err := ctx.Request().File("file")

					mockStorage, mockDriver, _ := mock.Storage()
					mockStorage.On("Disk", "local").Return(mockDriver).Once()
					mockDriver.On("PutFile", "test", fileInfo).Return("test/logo.png", nil).Once()
					mockStorage.On("Exists", "test/logo.png").Return(true).Once()

					if err != nil {
						ctx.Response().Success().String("get file error")
						return
					}
					filePath, err := fileInfo.Store("test")
					if err != nil {
						ctx.Response().Success().String("store file error: " + err.Error())
						return
					}

					extension, err := fileInfo.Extension()
					if err != nil {
						ctx.Response().Success().String("get file extension error: " + err.Error())
						return
					}

					ctx.Response().Success().Json(httpcontract.Json{
						"exist":              mockStorage.Exists(filePath),
						"hash_name_length":   len(fileInfo.HashName()),
						"hash_name_length1":  len(fileInfo.HashName("test")),
						"file_path_length":   len(filePath),
						"extension":          extension,
						"original_name":      fileInfo.GetClientOriginalName(),
						"original_extension": fileInfo.GetClientOriginalExtension(),
					})
				})

				payload := &bytes.Buffer{}
				writer := multipart.NewWriter(payload)
				logo, errFile1 := os.Open("../logo.png")
				defer logo.Close()
				part1, errFile1 := writer.CreateFormFile("file", filepath.Base("../logo.png"))
				_, errFile1 = io.Copy(part1, logo)
				if errFile1 != nil {
					return errFile1
				}
				err := writer.Close()
				if err != nil {
					return err
				}

				req, _ = http.NewRequest(method, url, payload)
				req.Header.Set("Content-Type", writer.FormDataContentType())

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"exist\":true,\"extension\":\"png\",\"file_path_length\":13,\"hash_name_length\":44,\"hash_name_length1\":49,\"original_extension\":\"png\",\"original_name\":\"logo.png\"}",
		},
		{
			name:   "GET with validator and validate pass",
			method: "GET",
			url:    "/validator/validate/success?name=Goravel",
			setup: func(method, url string) error {
				gin.Get("/validator/validate/success", func(ctx httpcontract.Context) {
					mockValication, _, _ := mock.Validation()
					mockValication.On("Rules").Return([]validation.Rule{}).Once()

					validator, err := ctx.Request().Validate(map[string]string{
						"name": "required",
					})
					if err != nil {
						ctx.Response().String(400, "Validate error: "+err.Error())
						return
					}
					if validator.Fails() {
						ctx.Response().String(400, fmt.Sprintf("Validate fail: %+v", validator.Errors().All()))
						return
					}

					type Test struct {
						Name string `form:"name" json:"name"`
					}
					var test Test
					if err := validator.Bind(&test); err != nil {
						ctx.Response().String(400, "Validate bind error: "+err.Error())
						return
					}

					ctx.Response().Success().Json(httpcontract.Json{
						"name": test.Name,
					})
				})
				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"name\":\"Goravel\"}",
		},
		{
			name:   "GET with validator but validate fail",
			method: "GET",
			url:    "/validator/validate/fail?name=Goravel",
			setup: func(method, url string) error {
				gin.Get("/validator/validate/fail", func(ctx httpcontract.Context) {
					mockValication, _, _ := mock.Validation()
					mockValication.On("Rules").Return([]validation.Rule{}).Once()

					validator, err := ctx.Request().Validate(map[string]string{
						"name1": "required",
					})
					if err != nil {
						ctx.Response().String(http.StatusBadRequest, "Validate error: "+err.Error())
						return
					}
					if validator.Fails() {
						ctx.Response().String(http.StatusBadRequest, fmt.Sprintf("Validate fail: %+v", validator.Errors().All()))
						return
					}

					ctx.Response().Success().Json(httpcontract.Json{
						"name": "",
					})
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusBadRequest,
			expectBody: "Validate fail: map[name1:map[required:name1 is required to not be empty]]",
		},
		{
			name:   "GET with validator and validate request pass",
			method: "GET",
			url:    "/validator/validate-request/success?name=Goravel",
			setup: func(method, url string) error {
				gin.Get("/validator/validate-request/success", func(ctx httpcontract.Context) {
					mockValication, _, _ := mock.Validation()
					mockValication.On("Rules").Return([]validation.Rule{}).Once()

					var createUser CreateUser
					validateErrors, err := ctx.Request().ValidateRequest(&createUser)
					if err != nil {
						ctx.Response().String(http.StatusBadRequest, "Validate error: "+err.Error())
						return
					}
					if validateErrors != nil {
						ctx.Response().String(http.StatusBadRequest, fmt.Sprintf("Validate fail: %+v", validateErrors.All()))
						return
					}

					ctx.Response().Success().Json(httpcontract.Json{
						"name": createUser.Name,
					})
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"name\":\"Goravel1\"}",
		},
		{
			name:   "GET with validator but validate request fail",
			method: "GET",
			url:    "/validator/validate-request/fail?name1=Goravel",
			setup: func(method, url string) error {
				gin.Get("/validator/validate-request/fail", func(ctx httpcontract.Context) {
					mockValication, _, _ := mock.Validation()
					mockValication.On("Rules").Return([]validation.Rule{}).Once()

					var createUser CreateUser
					validateErrors, err := ctx.Request().ValidateRequest(&createUser)
					if err != nil {
						ctx.Response().String(http.StatusBadRequest, "Validate error: "+err.Error())
						return
					}
					if validateErrors != nil {
						ctx.Response().String(http.StatusBadRequest, fmt.Sprintf("Validate fail: %+v", validateErrors.All()))
						return
					}

					ctx.Response().Success().Json(httpcontract.Json{
						"name": createUser.Name,
					})
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusBadRequest,
			expectBody: "Validate fail: map[name:map[required:name is required to not be empty]]",
		},
		{
			name:   "POST with validator and validate pass",
			method: "POST",
			url:    "/validator/validate/success",
			setup: func(method, url string) error {
				gin.Post("/validator/validate/success", func(ctx httpcontract.Context) {
					mockValication, _, _ := mock.Validation()
					mockValication.On("Rules").Return([]validation.Rule{}).Once()

					validator, err := ctx.Request().Validate(map[string]string{
						"name": "required",
					})
					if err != nil {
						ctx.Response().String(400, "Validate error: "+err.Error())
						return
					}
					if validator.Fails() {
						ctx.Response().String(400, fmt.Sprintf("Validate fail: %+v", validator.Errors().All()))
						return
					}

					type Test struct {
						Name string `form:"name" json:"name"`
					}
					var test Test
					if err := validator.Bind(&test); err != nil {
						ctx.Response().String(400, "Validate bind error: "+err.Error())
						return
					}

					ctx.Response().Success().Json(httpcontract.Json{
						"name": test.Name,
					})
				})

				payload := strings.NewReader(`{
					"name": "Goravel"
				}`)
				req, _ = http.NewRequest(method, url, payload)
				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"name\":\"Goravel\"}",
		},
		{
			name:   "POST with validator and validate fail",
			method: "POST",
			url:    "/validator/validate/fail",
			setup: func(method, url string) error {
				gin.Post("/validator/validate/fail", func(ctx httpcontract.Context) {
					mockValication, _, _ := mock.Validation()
					mockValication.On("Rules").Return([]validation.Rule{}).Once()

					validator, err := ctx.Request().Validate(map[string]string{
						"name1": "required",
					})
					if err != nil {
						ctx.Response().String(400, "Validate error: "+err.Error())
						return
					}
					if validator.Fails() {
						ctx.Response().String(400, fmt.Sprintf("Validate fail: %+v", validator.Errors().All()))
						return
					}

					ctx.Response().Success().Json(httpcontract.Json{
						"name": "",
					})
				})
				payload := strings.NewReader(`{
					"name": "Goravel"
				}`)
				req, _ = http.NewRequest(method, url, payload)
				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusBadRequest,
			expectBody: "Validate fail: map[name1:map[required:name1 is required to not be empty]]",
		},
		{
			name:   "POST with validator and validate request pass",
			method: "POST",
			url:    "/validator/validate-request/success",
			setup: func(method, url string) error {
				gin.Post("/validator/validate-request/success", func(ctx httpcontract.Context) {
					mockValication, _, _ := mock.Validation()
					mockValication.On("Rules").Return([]validation.Rule{}).Once()

					var createUser CreateUser
					validateErrors, err := ctx.Request().ValidateRequest(&createUser)
					if err != nil {
						ctx.Response().String(http.StatusBadRequest, "Validate error: "+err.Error())
						return
					}
					if validateErrors != nil {
						ctx.Response().String(http.StatusBadRequest, fmt.Sprintf("Validate fail: %+v", validateErrors.All()))
						return
					}

					ctx.Response().Success().Json(httpcontract.Json{
						"name": createUser.Name,
					})
				})

				payload := strings.NewReader(`{
					"name": "Goravel"
				}`)
				req, _ = http.NewRequest(method, url, payload)
				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"name\":\"Goravel1\"}",
		},
		{
			name:   "POST with validator and validate request fail",
			method: "POST",
			url:    "/validator/validate-request/fail",
			setup: func(method, url string) error {
				gin.Post("/validator/validate-request/fail", func(ctx httpcontract.Context) {
					mockValication, _, _ := mock.Validation()
					mockValication.On("Rules").Return([]validation.Rule{}).Once()

					var createUser CreateUser
					validateErrors, err := ctx.Request().ValidateRequest(&createUser)
					if err != nil {
						ctx.Response().String(http.StatusBadRequest, "Validate error: "+err.Error())
						return
					}
					if validateErrors != nil {
						ctx.Response().String(http.StatusBadRequest, fmt.Sprintf("Validate fail: %+v", validateErrors.All()))
						return
					}

					ctx.Response().Success().Json(httpcontract.Json{
						"name": createUser.Name,
					})
				})

				payload := strings.NewReader(`{
					"name1": "Goravel"
				}`)
				req, _ = http.NewRequest(method, url, payload)
				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusBadRequest,
			expectBody: "Validate fail: map[name:map[required:name is required to not be empty]]",
		},
		{
			name:   "POST with validator and validate request unauthorize",
			method: "POST",
			url:    "/validator/validate-request/unauthorize",
			setup: func(method, url string) error {
				gin.Post("/validator/validate-request/unauthorize", func(ctx httpcontract.Context) {
					var unauthorize Unauthorize
					validateErrors, err := ctx.Request().ValidateRequest(&unauthorize)
					if err != nil {
						ctx.Response().String(http.StatusBadRequest, "Validate error: "+err.Error())
						return
					}
					if validateErrors != nil {
						ctx.Response().String(http.StatusBadRequest, fmt.Sprintf("Validate fail: %+v", validateErrors.All()))
						return
					}

					ctx.Response().Success().Json(httpcontract.Json{
						"name": unauthorize.Name,
					})
				})
				payload := strings.NewReader(`{
					"name": "Goravel"
				}`)
				req, _ = http.NewRequest(method, url, payload)
				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusBadRequest,
			expectBody: "Validate error: error",
		},
	}

	for _, test := range tests {
		beforeEach()
		err := test.setup(test.method, test.url)
		assert.Nil(t, err)

		w := httptest.NewRecorder()
		gin.ServeHTTP(w, req)

		if test.expectBody != "" {
			assert.Equal(t, test.expectBody, w.Body.String(), test.name)
		}
		assert.Equal(t, test.expectCode, w.Code, test.name)
	}
}

func TestGinResponse(t *testing.T) {
	var (
		gin        *Gin
		req        *http.Request
		mockConfig *mockconfig.Config
	)
	beforeEach := func() {
		mockConfig = mock.Config()
		mockConfig.On("GetBool", "app.debug").Return(true).Once()

		gin = NewGin()
	}
	tests := []struct {
		name         string
		method       string
		url          string
		setup        func(method, url string) error
		expectCode   int
		expectBody   string
		expectHeader string
	}{
		{
			name:   "Json",
			method: "GET",
			url:    "/json",
			setup: func(method, url string) error {
				gin.Get("/json", func(ctx httpcontract.Context) {
					ctx.Response().Json(http.StatusOK, httpcontract.Json{
						"id": "1",
					})
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"1\"}",
		},
		{
			name:   "String",
			method: "GET",
			url:    "/string",
			setup: func(method, url string) error {
				gin.Get("/string", func(ctx httpcontract.Context) {
					ctx.Response().String(http.StatusCreated, "Goravel")
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusCreated,
			expectBody: "Goravel",
		},
		{
			name:   "Success Json",
			method: "GET",
			url:    "/success/json",
			setup: func(method, url string) error {
				gin.Get("/success/json", func(ctx httpcontract.Context) {
					ctx.Response().Success().Json(httpcontract.Json{
						"id": "1",
					})
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"1\"}",
		},
		{
			name:   "Success String",
			method: "GET",
			url:    "/success/string",
			setup: func(method, url string) error {
				gin.Get("/success/string", func(ctx httpcontract.Context) {
					ctx.Response().Success().String("Goravel")
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "Goravel",
		},
		{
			name:   "File",
			method: "GET",
			url:    "/file",
			setup: func(method, url string) error {
				gin.Get("/file", func(ctx httpcontract.Context) {
					ctx.Response().File("../logo.png")
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusOK,
		},
		{
			name:   "Download",
			method: "GET",
			url:    "/download",
			setup: func(method, url string) error {
				gin.Get("/download", func(ctx httpcontract.Context) {
					ctx.Response().Download("../logo.png", "1.png")
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusOK,
		},
		{
			name:   "Header",
			method: "GET",
			url:    "/header",
			setup: func(method, url string) error {
				gin.Get("/header", func(ctx httpcontract.Context) {
					ctx.Response().Header("Hello", "goravel").String(http.StatusOK, "Goravel")
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode:   http.StatusOK,
			expectBody:   "Goravel",
			expectHeader: "goravel",
		},
		{
			name:   "Origin",
			method: "GET",
			url:    "/origin",
			setup: func(method, url string) error {
				gin.GlobalMiddleware(func(ctx httpcontract.Context) {
					ctx.Response().Header("global", "goravel")
					ctx.Request().Next()

					assert.Equal(t, "Goravel", ctx.Response().Origin().Body().String())
					assert.Equal(t, "goravel", ctx.Response().Origin().Header().Get("global"))
					assert.Equal(t, 7, ctx.Response().Origin().Size())
					assert.Equal(t, 200, ctx.Response().Origin().Status())
				})
				gin.Get("/origin", func(ctx httpcontract.Context) {
					ctx.Response().String(http.StatusOK, "Goravel")
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "Goravel",
		},
	}

	for _, test := range tests {
		beforeEach()
		err := test.setup(test.method, test.url)
		assert.Nil(t, err)

		w := httptest.NewRecorder()
		gin.ServeHTTP(w, req)

		if test.expectBody != "" {
			assert.Equal(t, test.expectBody, w.Body.String(), test.name)
		}
		if test.expectHeader != "" {
			assert.Equal(t, test.expectHeader, strings.Join(w.Header().Values("Hello"), ""), test.name)
		}
		assert.Equal(t, test.expectCode, w.Code, test.name)
	}
}

type CreateUser struct {
	Name string `form:"name" json:"name"`
}

func (r *CreateUser) Authorize(ctx httpcontract.Context) error {
	return nil
}

func (r *CreateUser) Rules() map[string]string {
	return map[string]string{
		"name": "required",
	}
}

func (r *CreateUser) Messages() map[string]string {
	return map[string]string{}
}

func (r *CreateUser) Attributes() map[string]string {
	return map[string]string{}
}

func (r *CreateUser) PrepareForValidation(data validation.Data) {
	if name, exist := data.Get("name"); exist {
		_ = data.Set("name", name.(string)+"1")
	}
}

type Unauthorize struct {
	Name string `form:"name" json:"name"`
}

func (r *Unauthorize) Authorize(ctx httpcontract.Context) error {
	return errors.New("error")
}

func (r *Unauthorize) Rules() map[string]string {
	return map[string]string{
		"name": "required",
	}
}

func (r *Unauthorize) Messages() map[string]string {
	return map[string]string{}
}

func (r *Unauthorize) Attributes() map[string]string {
	return map[string]string{}
}

func (r *Unauthorize) PrepareForValidation(data validation.Data) {

}
