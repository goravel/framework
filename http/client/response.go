package client

import (
	"io"
	"net/http"
	"sync"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http/client"
)

type responseImpl struct {
	mu       sync.Mutex
	content  string
	decoded  map[string]any
	json     foundation.Json
	response *http.Response
}

func NewResponse(response *http.Response, json foundation.Json) client.Response {
	return &responseImpl{
		response: response,
		json:     json,
	}
}

func (r *responseImpl) Body() (string, error) {
	return r.getContent()
}

func (r *responseImpl) ClientError() bool {
	return r.getStatusCode() >= http.StatusBadRequest && r.getStatusCode() < http.StatusInternalServerError
}

func (r *responseImpl) Cookie(name string) *http.Cookie {
	return r.getCookie(name)
}

func (r *responseImpl) Cookies() []*http.Cookie {
	return r.response.Cookies()
}

func (r *responseImpl) Failed() bool {
	return r.ServerError() || r.ClientError()
}

func (r *responseImpl) Header(name string) string {
	return r.getHeader(name)
}

func (r *responseImpl) Headers() http.Header {
	return r.response.Header
}

func (r *responseImpl) Json() (map[string]any, error) {
	if r.decoded != nil {
		return r.decoded, nil
	}

	content, err := r.getContent()
	if err != nil {
		return nil, err
	}

	if err := r.json.Unmarshal([]byte(content), &r.decoded); err != nil {
		return nil, err
	}

	return r.decoded, nil
}

func (r *responseImpl) Redirect() bool {
	status := r.getStatusCode()
	return status >= http.StatusMultipleChoices && status < http.StatusBadRequest
}

func (r *responseImpl) ServerError() bool {
	return r.getStatusCode() >= http.StatusInternalServerError
}

func (r *responseImpl) Status() int {
	return r.getStatusCode()
}

func (r *responseImpl) Successful() bool {
	status := r.getStatusCode()
	return status >= http.StatusOK && status < http.StatusMultipleChoices
}

// OK checks if the status code is 200.
func (r *responseImpl) OK() bool {
	return r.getStatusCode() == http.StatusOK
}

// Created checks if the status code is 201.
func (r *responseImpl) Created() bool {
	return r.getStatusCode() == http.StatusCreated
}

// Accepted checks if the status code is 202.
func (r *responseImpl) Accepted() bool {
	return r.getStatusCode() == http.StatusAccepted
}

// NoContent checks if the status code is 204.
func (r *responseImpl) NoContent() bool {
	return r.getStatusCode() == http.StatusNoContent
}

// MovedPermanently checks if the status code is 301.
func (r *responseImpl) MovedPermanently() bool {
	return r.getStatusCode() == http.StatusMovedPermanently
}

// Found checks if the status code is 302.
func (r *responseImpl) Found() bool {
	return r.getStatusCode() == http.StatusFound
}

// BadRequest checks if the status code is 400.
func (r *responseImpl) BadRequest() bool {
	return r.getStatusCode() == http.StatusBadRequest
}

// Unauthorized checks if the status code is 401.
func (r *responseImpl) Unauthorized() bool {
	return r.getStatusCode() == http.StatusUnauthorized
}

// PaymentRequired checks if the status code is 402.
func (r *responseImpl) PaymentRequired() bool {
	return r.getStatusCode() == http.StatusPaymentRequired
}

// Forbidden checks if the status code is 403.
func (r *responseImpl) Forbidden() bool {
	return r.getStatusCode() == http.StatusForbidden
}

// NotFound checks if the status code is 404.
func (r *responseImpl) NotFound() bool {
	return r.getStatusCode() == http.StatusNotFound
}

// RequestTimeout checks if the status code is 408.
func (r *responseImpl) RequestTimeout() bool {
	return r.getStatusCode() == http.StatusRequestTimeout
}

// Conflict checks if the status code is 409.
func (r *responseImpl) Conflict() bool {
	return r.getStatusCode() == http.StatusConflict
}

// UnprocessableEntity checks if the status code is 422.
func (r *responseImpl) UnprocessableEntity() bool {
	return r.getStatusCode() == http.StatusUnprocessableEntity
}

// TooManyRequests checks if the status code is 429.
func (r *responseImpl) TooManyRequests() bool {
	return r.getStatusCode() == http.StatusTooManyRequests
}

func (r *responseImpl) getStatusCode() int {
	if r.response != nil {
		return r.response.StatusCode
	}
	return 0
}

func (r *responseImpl) getContent() (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.content != "" {
		return r.content, nil
	}

	defer r.response.Body.Close()

	content, err := io.ReadAll(r.response.Body)
	if err != nil {
		return "", err
	}

	r.content = string(content)
	return r.content, nil
}

func (r *responseImpl) getCookie(name string) *http.Cookie {
	for _, c := range r.response.Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func (r *responseImpl) getHeader(name string) string {
	return r.response.Header.Get(name)
}
