package client

import (
	"bytes"
	"io"
	"net/http"
	"os"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http/client"
)

var _ client.FakeResponse = (*FakeResponse)(nil)

type FakeResponse struct {
	json foundation.Json
}

func NewFakeResponse(json foundation.Json) *FakeResponse {
	return &FakeResponse{
		json: json,
	}
}

func (r *FakeResponse) Json(data any, status int) client.Response {
	content, err := r.json.Marshal(data)
	if err != nil {
		// Return error as body to help developer debug marshal issues in tests.
		return r.Make(err.Error(), http.StatusInternalServerError, nil)
	}

	return r.Make(string(content), status, map[string]string{
		"Content-Type": "application/json",
	})
}

func (r *FakeResponse) String(body string, status int) client.Response {
	return r.Make(body, status, nil)
}

func (r *FakeResponse) Status(code int) client.Response {
	return r.Make("", code, nil)
}

func (r *FakeResponse) OK() client.Response {
	return r.Status(http.StatusOK)
}

func (r *FakeResponse) File(path string, status int) client.Response {
	content, err := os.ReadFile(path)
	if err != nil {
		return r.Make("File not found: "+err.Error(), http.StatusInternalServerError, nil)
	}

	return r.Make(string(content), status, nil)
}

func (r *FakeResponse) Make(body string, status int, headers map[string]string) client.Response {
	resp := &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		// NopCloser prevents the body from being closed prematurely during testing.
		Body: io.NopCloser(bytes.NewBufferString(body)),
	}

	for key, value := range headers {
		resp.Header.Set(key, value)
	}

	return NewResponse(resp, r.json)
}
