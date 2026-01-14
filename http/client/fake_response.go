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

func (r *FakeResponse) File(path string, status int) client.Response {
	content, err := os.ReadFile(path)
	if err != nil {
		return r.make("Failed to read mock file "+path+": "+err.Error(), http.StatusInternalServerError, nil)
	}

	return r.make(string(content), status, nil)
}

func (r *FakeResponse) Json(data any, status int) client.Response {
	content, err := r.json.Marshal(data)
	if err != nil {
		return r.make("Failed to marshal mock JSON: "+err.Error(), http.StatusInternalServerError, nil)
	}

	header := http.Header{}
	header.Set("Content-Type", "application/json")

	return r.make(string(content), status, header)
}

func (r *FakeResponse) Make(body string, status int, header http.Header) client.Response {
	return r.make(body, status, header)
}

func (r *FakeResponse) OK() client.Response {
	return r.Status(http.StatusOK)
}

func (r *FakeResponse) Status(code int) client.Response {
	return r.make("", code, nil)
}

func (r *FakeResponse) String(body string, status int) client.Response {
	return r.make(body, status, nil)
}

func (r *FakeResponse) make(body string, status int, header http.Header) client.Response {
	resp := &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}

	for key, values := range header {
		for _, value := range values {
			resp.Header.Add(key, value)
		}
	}

	return NewResponse(resp, r.json)
}
