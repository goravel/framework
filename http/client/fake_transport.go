package client

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http/client"
)

const clientNameKey = "goravel_http_client_name"

var _ http.RoundTripper = (*FakeTransport)(nil)

type FakeTransport struct {
	mu       sync.Mutex
	recorded []client.Request
	rules    []*FakeRule
	json     foundation.Json
}

type FakeRule struct {
	pattern string
	regex   *regexp.Regexp
	handler func(client.Request) client.Response
}

func NewFakeTransport(j foundation.Json, mocks map[string]any) *FakeTransport {
	fakeTransport := &FakeTransport{
		json: j,
	}

	for p, v := range mocks {
		fakeTransport.rules = append(fakeTransport.rules, &FakeRule{
			pattern: p,
			regex:   fakeTransport.compilePattern(p),
			handler: fakeTransport.toHandler(v),
		})
	}

	sort.Slice(fakeTransport.rules, func(i, j int) bool {
		return len(fakeTransport.rules[i].pattern) > len(fakeTransport.rules[j].pattern)
	})

	return fakeTransport
}

func (r *FakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	mockReq := r.hydrate(req)

	r.mu.Lock()
	r.recorded = append(r.recorded, mockReq)
	r.mu.Unlock()

	handler := r.findHandler(req, mockReq.ClientName())
	if handler == nil {
		return nil, fmt.Errorf("goravel http fake: no mock defined for %s %s", req.Method, req.URL)
	}

	resp := handler(mockReq)
	if resp == nil {
		return nil, errors.New("goravel http fake: handler returned nil response")
	}

	if r, ok := resp.(*Response); ok {
		return r.response, nil
	}

	return nil, errors.New("goravel http fake: invalid response type")
}

func (r *FakeTransport) AssertSent(f func(client.Request) bool) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, r := range r.recorded {
		if f(r) {
			return true
		}
	}
	return false
}

func (r *FakeTransport) AssertSentCount(count int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.recorded) == count
}

func (r *FakeTransport) findHandler(req *http.Request, name string) func(client.Request) client.Response {
	url, path := req.URL.String(), req.URL.Path

	for _, rule := range r.rules {
		// Match by exact client name (e.g., "github").
		if !strings.ContainsAny(rule.pattern, "./:") && rule.pattern == name {
			return rule.handler
		}

		// Match by full URL pattern (e.g., "google.com/*").
		if rule.regex.MatchString(url) {
			return rule.handler
		}

		// Match by scoped client path (e.g., "github:/users/*").
		if name != "" && strings.HasPrefix(rule.pattern, name+":") && rule.regex.MatchString(path) {
			return rule.handler
		}
	}

	return nil
}

func (r *FakeTransport) toHandler(v any) func(client.Request) client.Response {
	switch h := v.(type) {
	case func(client.Request) client.Response:
		return h
	case client.Response:
		return func(_ client.Request) client.Response { return h }
	case string:
		return func(_ client.Request) client.Response { return NewResponseFactory(r.json).String(h, 200) }
	case int:
		return func(_ client.Request) client.Response { return NewResponseFactory(r.json).Status(h) }
	case *ResponseSequence:
		return func(_ client.Request) client.Response { return h.getNext() }
	default:
		return func(_ client.Request) client.Response { return NewResponseFactory(r.json).Status(200) }
	}
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

func (r *FakeTransport) compilePattern(p string) *regexp.Regexp {
	if i := strings.Index(p, ":"); i != -1 {
		p = p[i+1:]
	}
	if p == "*" {
		return regexp.MustCompile(".*")
	}
	expr := "^" + strings.ReplaceAll(regexp.QuoteMeta(p), "\\*", ".*") + "$"
	return regexp.MustCompile(expr)
}
