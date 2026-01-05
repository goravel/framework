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

// clientNameKey is the context key used to store/retrieve the client name.
const clientNameKey = "goravel_http_client_name"

var _ http.RoundTripper = (*FakeTransport)(nil)

type FakeTransport struct {
	mu       sync.Mutex
	rules    []*fakeRule
	recorded []client.Request
	json     foundation.Json
}

type fakeRule struct {
	original   string
	clientName string
	regex      *regexp.Regexp
	handler    func(client.Request) client.Response
}

func NewFakeTransport(json foundation.Json, mocks map[string]func(client.Request) client.Response) *FakeTransport {
	fakeTransport := &FakeTransport{
		recorded: make([]client.Request, 0),
		json:     json,
		rules:    make([]*fakeRule, 0, len(mocks)),
	}

	for pattern, handler := range mocks {
		fakeTransport.rules = append(fakeTransport.rules, compileRule(pattern, handler))
	}

	// This ensures that specific rules (e.g., "github.com/users/1") are checked
	// before broad wildcards (e.g., "github.com/*"), making matching deterministic.
	sort.Slice(fakeTransport.rules, func(i, j int) bool {
		return len(fakeTransport.rules[i].original) > len(fakeTransport.rules[j].original)
	})

	return fakeTransport
}

func (r *FakeTransport) RoundTrip(stdReq *http.Request) (*http.Response, error) {
	req := r.hydrateRequest(stdReq)
	r.mu.Lock()
	r.recorded = append(r.recorded, req)
	r.mu.Unlock()

	handler := r.match(stdReq.URL.String(), stdReq.URL.Path, req.ClientName())

	if handler == nil {
		return nil, fmt.Errorf("goravel http fake: no fake defined for request [%s] %s", stdReq.Method, stdReq.URL.String())
	}

	response := handler(req)
	if response == nil {
		return nil, errors.New("goravel http fake: handler returned nil response")
	}

	if casted, ok := response.(*Response); ok {
		return casted.response, nil
	}

	return nil, errors.New("goravel http fake: unknown response implementation")
}

func (r *FakeTransport) match(fullURL, path, clientName string) func(client.Request) client.Response {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, rule := range r.rules {
		if rule.clientName != "" {
			if rule.clientName != clientName {
				continue
			}
			// If the rule has no regex, it matches the entire client (e.g., "github").
			// If it has regex, it matches the path (e.g., "github:/users").
			if rule.regex == nil || rule.regex.MatchString(path) {
				return rule.handler
			}
			continue
		}

		if rule.regex != nil && rule.regex.MatchString(fullURL) {
			return rule.handler
		}
	}

	return nil
}

func (r *FakeTransport) hydrateRequest(httpRequest *http.Request) *Request {
	var bodyBytes []byte
	if httpRequest.Body != nil {
		// We ignore the error here as reading from a memory buffer is unlikely to fail.
		bodyBytes, _ = io.ReadAll(httpRequest.Body)
		// Reset the body immediately so it can be read again by downstream logic.
		httpRequest.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	clientName, _ := httpRequest.Context().Value(clientNameKey).(string)
	return &Request{
		json:        r.json,
		headers:     httpRequest.Header,
		cookies:     httpRequest.Cookies(),
		payloadBody: bodyBytes,
		method:      httpRequest.Method,
		fullUrl:     httpRequest.URL.String(),
		clientName:  clientName,
	}
}

func compileRule(pattern string, handler func(client.Request) client.Response) *fakeRule {
	rule := &fakeRule{
		original: pattern,
		handler:  handler,
	}

	if strings.Contains(pattern, ":") {
		parts := strings.SplitN(pattern, ":", 2)
		rule.clientName = parts[0]
		rule.regex = regexFromPattern(parts[1])
		return rule
	}

	if !strings.ContainsAny(pattern, "./") {
		rule.clientName = pattern
		return rule
	}

	rule.regex = regexFromPattern(pattern)
	return rule
}

func regexFromPattern(pattern string) *regexp.Regexp {
	if pattern == "*" {
		return regexp.MustCompile(".*")
	}

	// Escape strict characters to treat them literally (e.g., "." -> "\.")
	quote := regexp.QuoteMeta(pattern)
	// Convert the wildcard "*" back into the regex equivalent ".*"
	regexStr := strings.ReplaceAll(quote, "\\*", ".*")
	// Anchor the regex to ensure it matches the entire string
	regexStr = "^" + regexStr + "$"

	// We panic on compile error because this runs during test setup.
	// Invalid regex in a test setup should stop execution immediately.
	return regexp.MustCompile(regexStr)
}
