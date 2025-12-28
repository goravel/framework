package client

import (
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http/client"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/convert"
)

var _ client.Response = (*Response)(nil)

type Response struct {
	json     foundation.Json
	decoded  map[string]any
	response *http.Response
	content  string
	mu       sync.Mutex
}

func NewResponse(response *http.Response, json foundation.Json) *Response {
	return &Response{
		json:     json,
		response: response,
	}
}

func (r *Response) Bind(value any) error {
	content, err := r.getContent()
	if err != nil {
		return errors.HttpClientResponseBindFailed.Args(err)
	}

	if err := r.json.UnmarshalString(content, value); err != nil {
		return errors.HttpClientResponseUnmarshalFailed.Args(err)
	}

	return nil
}

func (r *Response) Body() (string, error) {
	return r.getContent()
}

func (r *Response) ClientError() bool {
	return r.getStatusCode() >= http.StatusBadRequest && r.getStatusCode() < http.StatusInternalServerError
}

func (r *Response) Cookie(name string) *http.Cookie {
	return r.getCookie(name)
}

func (r *Response) Cookies() []*http.Cookie {
	return r.response.Cookies()
}

func (r *Response) Failed() bool {
	return r.ServerError() || r.ClientError()
}

func (r *Response) Header(name string) string {
	return r.getHeader(name)
}

func (r *Response) Headers() http.Header {
	return r.response.Header
}

func (r *Response) Json() (map[string]any, error) {
	if r.decoded != nil {
		return r.decoded, nil
	}

	content, err := r.getContent()
	if err != nil {
		return nil, err
	}

	if err := r.json.UnmarshalString(content, &r.decoded); err != nil {
		return nil, err
	}

	return r.decoded, nil
}

func (r *Response) Redirect() bool {
	status := r.getStatusCode()
	return status >= http.StatusMultipleChoices && status < http.StatusBadRequest
}

func (r *Response) ServerError() bool {
	return r.getStatusCode() >= http.StatusInternalServerError
}

func (r *Response) Status() int {
	return r.getStatusCode()
}

func (r *Response) Stream() (io.ReadCloser, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// If the user already called Bind(), Body(), or Json(), the content is
	// stored in memory (r.content). We return a reader for this cached string
	// so the stream works seamlessly even after parsing.
	if r.content != "" {
		return io.NopCloser(strings.NewReader(r.content)), nil
	}

	if r.response == nil || r.response.Body == nil {
		return nil, errors.HttpClientResponseUnmarshalFailed.Args("response is nil")
	}

	// We give the raw network stream to the user.
	// Note: Calling Bind() after this point will likely fail or return empty
	// data because the stream will be consumed.
	return r.response.Body, nil
}

func (r *Response) Successful() bool {
	status := r.getStatusCode()
	return status >= http.StatusOK && status < http.StatusMultipleChoices
}

// OK checks if the status code is 200.
func (r *Response) OK() bool {
	return r.getStatusCode() == http.StatusOK
}

// Created checks if the status code is 201.
func (r *Response) Created() bool {
	return r.getStatusCode() == http.StatusCreated
}

// Accepted checks if the status code is 202.
func (r *Response) Accepted() bool {
	return r.getStatusCode() == http.StatusAccepted
}

// NoContent checks if the status code is 204.
func (r *Response) NoContent() bool {
	return r.getStatusCode() == http.StatusNoContent
}

// MovedPermanently checks if the status code is 301.
func (r *Response) MovedPermanently() bool {
	return r.getStatusCode() == http.StatusMovedPermanently
}

// Found checks if the status code is 302.
func (r *Response) Found() bool {
	return r.getStatusCode() == http.StatusFound
}

// BadRequest checks if the status code is 400.
func (r *Response) BadRequest() bool {
	return r.getStatusCode() == http.StatusBadRequest
}

// Unauthorized checks if the status code is 401.
func (r *Response) Unauthorized() bool {
	return r.getStatusCode() == http.StatusUnauthorized
}

// PaymentRequired checks if the status code is 402.
func (r *Response) PaymentRequired() bool {
	return r.getStatusCode() == http.StatusPaymentRequired
}

// Forbidden checks if the status code is 403.
func (r *Response) Forbidden() bool {
	return r.getStatusCode() == http.StatusForbidden
}

// NotFound checks if the status code is 404.
func (r *Response) NotFound() bool {
	return r.getStatusCode() == http.StatusNotFound
}

// RequestTimeout checks if the status code is 408.
func (r *Response) RequestTimeout() bool {
	return r.getStatusCode() == http.StatusRequestTimeout
}

// Conflict checks if the status code is 409.
func (r *Response) Conflict() bool {
	return r.getStatusCode() == http.StatusConflict
}

// UnprocessableEntity checks if the status code is 422.
func (r *Response) UnprocessableEntity() bool {
	return r.getStatusCode() == http.StatusUnprocessableEntity
}

// TooManyRequests checks if the status code is 429.
func (r *Response) TooManyRequests() bool {
	return r.getStatusCode() == http.StatusTooManyRequests
}

func (r *Response) getStatusCode() int {
	if r.response != nil {
		return r.response.StatusCode
	}
	return 0
}

func (r *Response) getContent() (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.content != "" {
		return r.content, nil
	}

	defer errors.Ignore(r.response.Body.Close)

	content, err := io.ReadAll(r.response.Body)
	if err != nil {
		return "", err
	}

	r.content = convert.UnsafeString(content)
	return r.content, nil
}

func (r *Response) getCookie(name string) *http.Cookie {
	for _, c := range r.response.Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func (r *Response) getHeader(name string) string {
	return r.response.Header.Get(name)
}
