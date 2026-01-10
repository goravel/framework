package client

import (
	"bytes"
	"io"
	"net/http"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
)

type contextKey string

const clientNameKey contextKey = "goravel_http_client_name"

var _ http.RoundTripper = (*FakeTransport)(nil)

type FakeTransport struct {
	state *FakeState
	base  http.RoundTripper
	json  foundation.Json
}

func NewFakeTransport(state *FakeState, base http.RoundTripper, json foundation.Json) *FakeTransport {
	if base == nil {
		base = http.DefaultTransport
	}
	return &FakeTransport{
		state: state,
		base:  base,
		json:  json,
	}
}

func (r *FakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	mockReq := r.hydrate(req)
	r.state.Record(mockReq)

	handler := r.state.Match(req, mockReq.ClientName())
	if handler == nil {
		if r.state.ShouldPreventStray(req.URL.String()) {
			return nil, errors.HttpClientStrayRequest.Args(req.Method, req.URL.String())
		}

		return r.base.RoundTrip(req)
	}

	resp := handler(mockReq)
	if resp == nil {
		return nil, errors.HttpClientHandlerReturnedNil
	}
	return resp.Origin(), nil
}

func (r *FakeTransport) hydrate(req *http.Request) *Request {
	var body []byte
	if req.Body != nil {
		body, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	name, _ := req.Context().Value(clientNameKey).(string)

	return &Request{
		json:        r.json,
		headers:     req.Header,
		cookies:     req.Cookies(),
		payloadBody: body,
		method:      req.Method,
		fullUrl:     req.URL.String(),
		clientName:  name,
		queryParams: req.URL.Query(),
		urlParams:   make(map[string]string),
	}
}
