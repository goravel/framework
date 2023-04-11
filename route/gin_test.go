package route

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	configmocks "github.com/goravel/framework/contracts/config/mocks"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/testing/mock"
)

func TestFallback(t *testing.T) {
	var (
		gin        *Gin
		mockConfig *configmocks.Config
	)
	beforeEach := func() {
		mockConfig = mock.Config()
		mockConfig.On("GetBool", "app.debug").Return(true).Once()

		gin = NewGin()
	}
	tests := []struct {
		name       string
		setup      func(req *http.Request)
		method     string
		url        string
		expectCode int
		expectBody string
	}{
		{
			name: "success",
			setup: func(req *http.Request) {
				gin.Fallback(func(ctx contractshttp.Context) {
					ctx.Response().String(404, "not found")
				})
			},
			method:     "GET",
			url:        "/fallback",
			expectCode: http.StatusNotFound,
			expectBody: "not found",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(test.method, test.url, nil)
			if test.setup != nil {
				test.setup(req)
			}
			gin.ServeHTTP(w, req)

			if test.expectBody != "" {
				assert.Equal(t, test.expectBody, w.Body.String(), test.name)
			}
			assert.Equal(t, test.expectCode, w.Code, test.name)
		})
	}
}

func TestRun(t *testing.T) {
	var mockConfig *configmocks.Config
	var route *Gin

	tests := []struct {
		name        string
		setup       func(host string, port string) error
		host        string
		port        string
		expectError error
	}{
		{
			name: "error when default host is empty",
			setup: func(host string, port string) error {
				mockConfig.On("GetString", "http.host").Return(host).Once()

				go func() {
					assert.EqualError(t, route.Run(), "host can't be empty")
				}()
				time.Sleep(1 * time.Second)

				return errors.New("error")
			},
		},
		{
			name: "error when default port is empty",
			setup: func(host string, port string) error {
				mockConfig.On("GetString", "http.host").Return(host).Once()
				mockConfig.On("GetString", "http.port").Return(port).Once()

				go func() {
					assert.EqualError(t, route.Run(), "port can't be empty")
				}()
				time.Sleep(1 * time.Second)

				return errors.New("error")
			},
			host: "127.0.0.1",
		},
		{
			name: "use default host",
			setup: func(host string, port string) error {
				mockConfig.On("GetBool", "app.debug").Return(true).Once()
				mockConfig.On("GetString", "http.host").Return(host).Once()
				mockConfig.On("GetString", "http.port").Return(port).Once()

				go func() {
					assert.Nil(t, route.Run())
				}()

				return nil
			},
			host: "127.0.0.1",
			port: "3001",
		},
		{
			name: "use custom host",
			setup: func(host string, port string) error {
				mockConfig.On("GetBool", "app.debug").Return(true).Once()

				go func() {
					assert.Nil(t, route.Run(host))
				}()

				return nil
			},
			host: "127.0.0.1:3002",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockConfig = mock.Config()
			mockConfig.On("GetBool", "app.debug").Return(true).Once()
			route = NewGin()
			route.Get("/", func(ctx contractshttp.Context) {
				ctx.Response().Json(200, contractshttp.Json{
					"Hello": "Goravel",
				})
			})
			if err := test.setup(test.host, test.port); err == nil {
				time.Sleep(1 * time.Second)
				hostUrl := "http://" + test.host
				if test.port != "" {
					hostUrl = hostUrl + ":" + test.port
				}
				resp, err := http.Get(hostUrl)
				assert.Nil(t, err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				assert.Nil(t, err)
				assert.Equal(t, "{\"Hello\":\"Goravel\"}", string(body))
			}
			mockConfig.AssertExpectations(t)
		})
	}
}

func TestRunTLS(t *testing.T) {
	var mockConfig *configmocks.Config
	var route *Gin

	tests := []struct {
		name        string
		setup       func(host string, port string) error
		host        string
		port        string
		expectError error
	}{
		{
			name: "error when default host is empty",
			setup: func(host string, port string) error {
				mockConfig.On("GetString", "http.tls.host").Return(host).Once()

				go func() {
					assert.EqualError(t, route.RunTLS(), "host can't be empty")
				}()
				time.Sleep(1 * time.Second)

				return errors.New("error")
			},
		},
		{
			name: "error when default port is empty",
			setup: func(host string, port string) error {
				mockConfig.On("GetString", "http.tls.host").Return(host).Once()
				mockConfig.On("GetString", "http.tls.port").Return(port).Once()

				go func() {
					assert.EqualError(t, route.RunTLS(), "port can't be empty")
				}()
				time.Sleep(1 * time.Second)

				return errors.New("error")
			},
			host: "127.0.0.1",
		},
		{
			name: "use default host",
			setup: func(host string, port string) error {
				mockConfig.On("GetBool", "app.debug").Return(true).Once()
				mockConfig.On("GetString", "http.tls.host").Return(host).Once()
				mockConfig.On("GetString", "http.tls.port").Return(port).Once()
				mockConfig.On("GetString", "http.tls.ssl.cert").Return("test_ca.crt").Once()
				mockConfig.On("GetString", "http.tls.ssl.key").Return("test_ca.key").Once()

				go func() {
					assert.Nil(t, route.RunTLS())
				}()

				return nil
			},
			host: "127.0.0.1",
			port: "3003",
		},
		{
			name: "use custom host",
			setup: func(host string, port string) error {
				mockConfig.On("GetBool", "app.debug").Return(true).Once()
				mockConfig.On("GetString", "http.tls.ssl.cert").Return("test_ca.crt").Once()
				mockConfig.On("GetString", "http.tls.ssl.key").Return("test_ca.key").Once()

				go func() {
					assert.Nil(t, route.RunTLS(host))
				}()

				return nil
			},
			host: "127.0.0.1:3004",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockConfig = mock.Config()
			mockConfig.On("GetBool", "app.debug").Return(true).Once()
			route = NewGin()
			route.Get("/", func(ctx contractshttp.Context) {
				ctx.Response().Json(200, contractshttp.Json{
					"Hello": "Goravel",
				})
			})
			if err := test.setup(test.host, test.port); err == nil {
				time.Sleep(1 * time.Second)
				tr := &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				}
				client := &http.Client{Transport: tr}
				hostUrl := "https://" + test.host
				if test.port != "" {
					hostUrl = hostUrl + ":" + test.port
				}
				resp, err := client.Get(hostUrl)
				assert.Nil(t, err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				assert.Nil(t, err)
				assert.Equal(t, "{\"Hello\":\"Goravel\"}", string(body))
			}
			mockConfig.AssertExpectations(t)
		})
	}
}

func TestRunTLSWithCert(t *testing.T) {
	var mockConfig *configmocks.Config
	var route *Gin

	tests := []struct {
		name        string
		setup       func(host string) error
		host        string
		expectError error
	}{
		{
			name: "error when default host is empty",
			setup: func(host string) error {
				go func() {
					assert.EqualError(t, route.RunTLSWithCert(host, "test_ca.crt", "test_ca.key"), "host can't be empty")
				}()
				time.Sleep(1 * time.Second)

				return errors.New("error")
			},
		},
		{
			name: "use default host",
			setup: func(host string) error {
				mockConfig.On("GetBool", "app.debug").Return(true).Once()

				go func() {
					assert.Nil(t, route.RunTLSWithCert(host, "test_ca.crt", "test_ca.key"))
				}()

				return nil
			},
			host: "127.0.0.1:3005",
		},
		{
			name: "use custom host",
			setup: func(host string) error {
				mockConfig.On("GetBool", "app.debug").Return(true).Once()

				go func() {
					assert.Nil(t, route.RunTLSWithCert(host, "test_ca.crt", "test_ca.key"))
				}()

				return nil
			},
			host: "127.0.0.1:3006",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockConfig = mock.Config()
			mockConfig.On("GetBool", "app.debug").Return(true).Once()
			route = NewGin()
			route.Get("/", func(ctx contractshttp.Context) {
				ctx.Response().Json(200, contractshttp.Json{
					"Hello": "Goravel",
				})
			})
			if err := test.setup(test.host); err == nil {
				time.Sleep(1 * time.Second)
				tr := &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				}
				client := &http.Client{Transport: tr}
				resp, err := client.Get("https://" + test.host)
				assert.Nil(t, err)
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				assert.Nil(t, err)
				assert.Equal(t, "{\"Hello\":\"Goravel\"}", string(body))
			}
			mockConfig.AssertExpectations(t)
		})
	}
}

func TestGinRequest(t *testing.T) {
	var (
		gin        *Gin
		req        *http.Request
		mockConfig *configmocks.Config
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
			url:    "/methods/1?name=Goravel",
			setup: func(method, url string) error {
				gin.Get("/methods/{id}", func(ctx contractshttp.Context) {
					ctx.Response().Success().Json(contractshttp.Json{
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
			expectBody: "{\"full_url\":\"\",\"header\":\"goravel\",\"id\":\"1\",\"ip\":\"\",\"method\":\"GET\",\"name\":\"Goravel\",\"path\":\"/methods/1\",\"url\":\"\"}",
		},
		{
			name:   "Headers",
			method: "GET",
			url:    "/headers",
			setup: func(method, url string) error {
				gin.Get("/headers", func(ctx contractshttp.Context) {
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
			name:   "Route",
			method: "GET",
			url:    "/route/1/2/3/a",
			setup: func(method, url string) error {
				gin.Get("/route/{string}/{int}/{int64}/{string1}", func(ctx contractshttp.Context) {
					ctx.Response().Success().Json(contractshttp.Json{
						"string": ctx.Request().Route("string"),
						"int":    ctx.Request().RouteInt("int"),
						"int64":  ctx.Request().RouteInt64("int64"),
						"error":  ctx.Request().RouteInt("string1"),
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
			expectBody: "{\"error\":0,\"int\":2,\"int64\":3,\"string\":\"1\"}",
		},
		{
			name:   "Input - from json",
			method: "POST",
			url:    "/input1/1?id=2",
			setup: func(method, url string) error {
				gin.Post("/input1/{id}", func(ctx contractshttp.Context) {
					ctx.Response().Success().Json(contractshttp.Json{
						"id": ctx.Request().Input("id"),
					})
				})

				payload := strings.NewReader(`{
					"id": "3"
				}`)
				req, _ = http.NewRequest(method, url, payload)
				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"3\"}",
		},
		{
			name:   "Input - from form",
			method: "POST",
			url:    "/input2/1?id=2",
			setup: func(method, url string) error {
				gin.Post("/input2/{id}", func(ctx contractshttp.Context) {
					ctx.Response().Success().Json(contractshttp.Json{
						"id": ctx.Request().Input("id"),
					})
				})

				payload := &bytes.Buffer{}
				writer := multipart.NewWriter(payload)
				if err := writer.WriteField("id", "4"); err != nil {
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
			expectBody: "{\"id\":\"4\"}",
		},
		{
			name:   "Input - from query",
			method: "POST",
			url:    "/input3/1?id=2",
			setup: func(method, url string) error {
				gin.Post("/input3/{id}", func(ctx contractshttp.Context) {
					ctx.Response().Success().Json(contractshttp.Json{
						"id": ctx.Request().Input("id"),
					})
				})

				req, _ = http.NewRequest(method, url, nil)
				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"2\"}",
		},
		{
			name:   "Input - from route",
			method: "POST",
			url:    "/input4/1",
			setup: func(method, url string) error {
				gin.Post("/input4/{id}", func(ctx contractshttp.Context) {
					ctx.Response().Success().Json(contractshttp.Json{
						"id": ctx.Request().Input("id"),
					})
				})

				req, _ = http.NewRequest(method, url, nil)
				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"1\"}",
		},
		{
			name:   "Input - empty",
			method: "POST",
			url:    "/input5/1",
			setup: func(method, url string) error {
				gin.Post("/input5/{id}", func(ctx contractshttp.Context) {
					ctx.Response().Success().Json(contractshttp.Json{
						"id1": ctx.Request().Input("id1"),
					})
				})

				req, _ = http.NewRequest(method, url, nil)
				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"id1\":\"\"}",
		},
		{
			name:   "Input - default",
			method: "POST",
			url:    "/input6/1",
			setup: func(method, url string) error {
				gin.Post("/input6/{id}", func(ctx contractshttp.Context) {
					ctx.Response().Success().Json(contractshttp.Json{
						"id1": ctx.Request().Input("id1", "2"),
					})
				})

				req, _ = http.NewRequest(method, url, nil)
				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"id1\":\"2\"}",
		},
		{
			name:   "InputInt",
			method: "POST",
			url:    "/input-int/1",
			setup: func(method, url string) error {
				gin.Post("/input-int/{id}", func(ctx contractshttp.Context) {
					ctx.Response().Success().Json(contractshttp.Json{
						"id": ctx.Request().InputInt("id"),
					})
				})

				req, _ = http.NewRequest(method, url, nil)
				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"id\":1}",
		},
		{
			name:   "InputInt64",
			method: "POST",
			url:    "/input-int64/1",
			setup: func(method, url string) error {
				gin.Post("/input-int64/{id}", func(ctx contractshttp.Context) {
					ctx.Response().Success().Json(contractshttp.Json{
						"id": ctx.Request().InputInt64("id"),
					})
				})

				req, _ = http.NewRequest(method, url, nil)
				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"id\":1}",
		},
		{
			name:   "InputBool",
			method: "POST",
			url:    "/input-bool/1/true/on/yes/a",
			setup: func(method, url string) error {
				gin.Post("/input-bool/{id1}/{id2}/{id3}/{id4}/{id5}", func(ctx contractshttp.Context) {
					ctx.Response().Success().Json(contractshttp.Json{
						"id1": ctx.Request().InputBool("id1"),
						"id2": ctx.Request().InputBool("id2"),
						"id3": ctx.Request().InputBool("id3"),
						"id4": ctx.Request().InputBool("id4"),
						"id5": ctx.Request().InputBool("id5"),
					})
				})

				req, _ = http.NewRequest(method, url, nil)
				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"id1\":true,\"id2\":true,\"id3\":true,\"id4\":true,\"id5\":false}",
		},
		{
			name:   "Form",
			method: "POST",
			url:    "/form",
			setup: func(method, url string) error {
				gin.Post("/form", func(ctx contractshttp.Context) {
					ctx.Response().Success().Json(contractshttp.Json{
						"name":  ctx.Request().Form("name", "Hello"),
						"name1": ctx.Request().Form("name1", "Hello"),
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
			expectBody: "{\"name\":\"Goravel\",\"name1\":\"Hello\"}",
		},
		{
			name:   "Json",
			method: "POST",
			url:    "/json",
			setup: func(method, url string) error {
				gin.Post("/json", func(ctx contractshttp.Context) {
					ctx.Response().Success().Json(contractshttp.Json{
						"name":   ctx.Request().Json("name"),
						"info":   ctx.Request().Json("info"),
						"avatar": ctx.Request().Json("avatar", "logo"),
					})
				})

				payload := strings.NewReader(`{
					"name": "Goravel",
					"info": {"avatar": "logo"}
				}`)
				req, _ = http.NewRequest(method, url, payload)
				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"avatar\":\"logo\",\"info\":\"\",\"name\":\"Goravel\"}",
		},
		{
			name:   "Bind",
			method: "POST",
			url:    "/bind",
			setup: func(method, url string) error {
				gin.Post("/bind", func(ctx contractshttp.Context) {
					type Test struct {
						Name string
					}
					var test Test
					_ = ctx.Request().Bind(&test)
					ctx.Response().Success().Json(contractshttp.Json{
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
			name:   "Query",
			method: "GET",
			url:    "/query?string=Goravel&int=1&int64=2&bool1=1&bool2=true&bool3=on&bool4=yes&bool5=0&error=a",
			setup: func(method, url string) error {
				gin.Get("/query", func(ctx contractshttp.Context) {
					ctx.Response().Success().Json(contractshttp.Json{
						"string":        ctx.Request().Query("string", ""),
						"int":           ctx.Request().QueryInt("int", 11),
						"int_default":   ctx.Request().QueryInt("int_default", 11),
						"int64":         ctx.Request().QueryInt64("int64", 22),
						"int64_default": ctx.Request().QueryInt64("int64_default", 22),
						"bool1":         ctx.Request().QueryBool("bool1"),
						"bool2":         ctx.Request().QueryBool("bool2"),
						"bool3":         ctx.Request().QueryBool("bool3"),
						"bool4":         ctx.Request().QueryBool("bool4"),
						"bool5":         ctx.Request().QueryBool("bool5"),
						"error":         ctx.Request().QueryInt("error", 33),
					})
				})

				req, _ = http.NewRequest(method, url, nil)
				req.Header.Set("Content-Type", "application/json")

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "{\"bool1\":true,\"bool2\":true,\"bool3\":true,\"bool4\":true,\"bool5\":false,\"error\":0,\"int\":1,\"int64\":2,\"int64_default\":22,\"int_default\":11,\"string\":\"Goravel\"}",
		},
		{
			name:   "QueryArray",
			method: "GET",
			url:    "/query-array?name=Goravel&name=Goravel1",
			setup: func(method, url string) error {
				gin.Get("/query-array", func(ctx contractshttp.Context) {
					ctx.Response().Success().Json(contractshttp.Json{
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
			url:    "/query-map?name[a]=Goravel&name[b]=Goravel1",
			setup: func(method, url string) error {
				gin.Get("/query-map", func(ctx contractshttp.Context) {
					ctx.Response().Success().Json(contractshttp.Json{
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
				gin.Post("/file", func(ctx contractshttp.Context) {
					mockConfig.On("GetString", "app.name").Return("goravel").Once()
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

					ctx.Response().Success().Json(contractshttp.Json{
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
				logo, err := os.Open("../logo.png")
				if err != nil {
					return err
				}
				defer logo.Close()
				part1, err := writer.CreateFormFile("file", filepath.Base("../logo.png"))
				if err != nil {
					return err
				}

				if _, err = io.Copy(part1, logo); err != nil {
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
			expectBody: "{\"exist\":true,\"extension\":\"png\",\"file_path_length\":13,\"hash_name_length\":44,\"hash_name_length1\":49,\"original_extension\":\"png\",\"original_name\":\"logo.png\"}",
		},
		{
			name:   "GET with validator and validate pass",
			method: "GET",
			url:    "/validator/validate/success?name=Goravel",
			setup: func(method, url string) error {
				gin.Get("/validator/validate/success", func(ctx contractshttp.Context) {
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

					ctx.Response().Success().Json(contractshttp.Json{
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
				gin.Get("/validator/validate/fail", func(ctx contractshttp.Context) {
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

					ctx.Response().Success().Json(contractshttp.Json{
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
				gin.Get("/validator/validate-request/success", func(ctx contractshttp.Context) {
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

					ctx.Response().Success().Json(contractshttp.Json{
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
				gin.Get("/validator/validate-request/fail", func(ctx contractshttp.Context) {
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

					ctx.Response().Success().Json(contractshttp.Json{
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
				gin.Post("/validator/validate/success", func(ctx contractshttp.Context) {
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

					ctx.Response().Success().Json(contractshttp.Json{
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
				gin.Post("/validator/validate/fail", func(ctx contractshttp.Context) {
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

					ctx.Response().Success().Json(contractshttp.Json{
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
				gin.Post("/validator/validate-request/success", func(ctx contractshttp.Context) {
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

					ctx.Response().Success().Json(contractshttp.Json{
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
				gin.Post("/validator/validate-request/fail", func(ctx contractshttp.Context) {
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

					ctx.Response().Success().Json(contractshttp.Json{
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
				gin.Post("/validator/validate-request/unauthorize", func(ctx contractshttp.Context) {
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

					ctx.Response().Success().Json(contractshttp.Json{
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
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			err := test.setup(test.method, test.url)
			assert.Nil(t, err)

			w := httptest.NewRecorder()
			gin.ServeHTTP(w, req)

			if test.expectBody != "" {
				assert.Equal(t, test.expectBody, w.Body.String(), test.name)
			}
			assert.Equal(t, test.expectCode, w.Code)
		})
	}
}

func TestGinResponse(t *testing.T) {
	var (
		gin        *Gin
		req        *http.Request
		mockConfig *configmocks.Config
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
			name:   "Data",
			method: "GET",
			url:    "/data",
			setup: func(method, url string) error {
				gin.Get("/data", func(ctx contractshttp.Context) {
					ctx.Response().Data(http.StatusOK, "text/html; charset=utf-8", []byte("<b>Goravel</b>"))
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "<b>Goravel</b>",
		},
		{
			name:   "Success Data",
			method: "GET",
			url:    "/success/data",
			setup: func(method, url string) error {
				gin.Get("/success/data", func(ctx contractshttp.Context) {
					ctx.Response().Success().Data("text/html; charset=utf-8", []byte("<b>Goravel</b>"))
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusOK,
			expectBody: "<b>Goravel</b>",
		},
		{
			name:   "Json",
			method: "GET",
			url:    "/json",
			setup: func(method, url string) error {
				gin.Get("/json", func(ctx contractshttp.Context) {
					ctx.Response().Json(http.StatusOK, contractshttp.Json{
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
				gin.Get("/string", func(ctx contractshttp.Context) {
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
				gin.Get("/success/json", func(ctx contractshttp.Context) {
					ctx.Response().Success().Json(contractshttp.Json{
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
				gin.Get("/success/string", func(ctx contractshttp.Context) {
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
				gin.Get("/file", func(ctx contractshttp.Context) {
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
				gin.Get("/download", func(ctx contractshttp.Context) {
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
				gin.Get("/header", func(ctx contractshttp.Context) {
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
				gin.GlobalMiddleware(func(ctx contractshttp.Context) {
					ctx.Response().Header("global", "goravel")
					ctx.Request().Next()

					assert.Equal(t, "Goravel", ctx.Response().Origin().Body().String())
					assert.Equal(t, "goravel", ctx.Response().Origin().Header().Get("global"))
					assert.Equal(t, 7, ctx.Response().Origin().Size())
					assert.Equal(t, 200, ctx.Response().Origin().Status())
				})
				gin.Get("/origin", func(ctx contractshttp.Context) {
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
		{
			name:   "Redirect",
			method: "GET",
			url:    "/redirect",
			setup: func(method, url string) error {
				gin.Get("/redirect", func(ctx contractshttp.Context) {
					ctx.Response().Redirect(http.StatusMovedPermanently, "https://goravel.dev")
				})

				var err error
				req, err = http.NewRequest(method, url, nil)
				if err != nil {
					return err
				}

				return nil
			},
			expectCode: http.StatusMovedPermanently,
			expectBody: "<a href=\"https://goravel.dev\">Moved Permanently</a>.\n\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
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
		})
	}
}

type CreateUser struct {
	Name string `form:"name" json:"name"`
}

func (r *CreateUser) Authorize(ctx contractshttp.Context) error {
	return nil
}

func (r *CreateUser) Rules(ctx contractshttp.Context) map[string]string {
	return map[string]string{
		"name": "required",
	}
}

func (r *CreateUser) Messages(ctx contractshttp.Context) map[string]string {
	return map[string]string{}
}

func (r *CreateUser) Attributes(ctx contractshttp.Context) map[string]string {
	return map[string]string{}
}

func (r *CreateUser) PrepareForValidation(ctx contractshttp.Context, data validation.Data) error {
	if name, exist := data.Get("name"); exist {
		return data.Set("name", name.(string)+"1")
	}

	return nil
}

type Unauthorize struct {
	Name string `form:"name" json:"name"`
}

func (r *Unauthorize) Authorize(ctx contractshttp.Context) error {
	return errors.New("error")
}

func (r *Unauthorize) Rules(ctx contractshttp.Context) map[string]string {
	return map[string]string{
		"name": "required",
	}
}

func (r *Unauthorize) Messages(ctx contractshttp.Context) map[string]string {
	return map[string]string{}
}

func (r *Unauthorize) Attributes(ctx contractshttp.Context) map[string]string {
	return map[string]string{}
}

func (r *Unauthorize) PrepareForValidation(ctx contractshttp.Context, data validation.Data) error {
	return nil
}
