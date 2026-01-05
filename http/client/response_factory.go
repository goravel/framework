package client

import (
	"bytes"
	"io"
	"net/http"
	"os"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http/client"
)

var _ client.ResponseFactory = (*ResponseFactory)(nil)

type ResponseFactory struct {
	json foundation.Json
}

func NewResponseFactory(json foundation.Json) *ResponseFactory {
	return &ResponseFactory{
		json: json,
	}
}

func (r *ResponseFactory) Json(data any, status int) client.Response {
	// If the data cannot be marshaled, we return an empty JSON object to avoid panic.
	// In a test environment, valid structs are expected.
	content, err := r.json.Marshal(data)
	if err != nil {
		return r.Make("{}", http.StatusInternalServerError, nil)
	}

	return r.Make(string(content), status, map[string]string{
		"Content-Type": "application/json",
	})
}

func (r *ResponseFactory) String(body string, status int) client.Response {
	return r.Make(body, status, nil)
}

func (r *ResponseFactory) Status(code int) client.Response {
	return r.Make("", code, nil)
}

func (r *ResponseFactory) Success() client.Response {
	return r.Status(http.StatusOK)
}

func (r *ResponseFactory) File(path string, status int) client.Response {
	content, err := os.ReadFile(path)
	if err != nil {
		// If the file is missing during a test setup, we treat it as a setup error.
		// Returning a 404 or 500 helps the developer debug that the mock file is missing.
		return r.String("Error reading fake file: "+err.Error(), http.StatusInternalServerError)
	}

	return r.Make(string(content), status, nil)
}

func (r *ResponseFactory) Make(body string, status int, headers map[string]string) client.Response {
	httpResp := &http.Response{
		StatusCode: status,
		// We use NopCloser to create a readable stream from the string body
		Body:   io.NopCloser(bytes.NewBufferString(body)),
		Header: make(http.Header),
	}

	for key, value := range headers {
		httpResp.Header.Set(key, value)
	}

	return NewResponse(httpResp, r.json)
}
