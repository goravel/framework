package client

import (
	"io"
	"net/http"
	"sync"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http/client"
)

var _ client.Response = &responseImpl{}

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
	return r.getStatusCode() >= 400 && r.getStatusCode() < 500
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

func (r *responseImpl) Headers() http.Header {
	return r.response.Header
}

func (r *responseImpl) Json() (map[string]any, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

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
	return r.getStatusCode() >= 300 && r.getStatusCode() < 400
}

func (r *responseImpl) ServerError() bool {
	return r.getStatusCode() >= 500
}

func (r *responseImpl) Status() int {
	return r.getStatusCode()
}

func (r *responseImpl) Successful() bool {
	return r.getStatusCode() >= 200 && r.getStatusCode() < 300
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

	if r.response.Body == nil {
		return "", io.EOF
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
