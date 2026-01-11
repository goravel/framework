package client

import (
	"net/http"
	"regexp"
	"sort"
	"sync"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http/client"
)

type FakeState struct {
	mu                   sync.RWMutex
	recorded             []client.Request
	rules                []*FakeRule
	allowedStrayPatterns []*regexp.Regexp
	preventStrayRequests bool
}

func NewFakeState(json foundation.Json, mocks map[string]any) *FakeState {
	rules := make([]*FakeRule, 0, len(mocks))
	for p, v := range mocks {
		rules = append(rules, NewFakeRule(p, toHandler(json, v)))
	}

	sort.Slice(rules, func(i, j int) bool {
		return len(rules[i].pattern) > len(rules[j].pattern)
	})

	return &FakeState{
		rules: rules,
	}
}

func (r *FakeState) Record(req client.Request) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.recorded = append(r.recorded, req)
}

func (r *FakeState) Match(req *http.Request, name string) func(client.Request) client.Response {
	for _, rule := range r.rules {
		if rule.Matches(req, name) {
			return rule.handler
		}
	}
	return nil
}

func (r *FakeState) ShouldPreventStray(url string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.preventStrayRequests {
		return false
	}

	for _, p := range r.allowedStrayPatterns {
		if p.MatchString(url) {
			return false
		}
	}
	return true
}

func (r *FakeState) PreventStrayRequests() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.preventStrayRequests = true
}

func (r *FakeState) AllowStrayRequests(patterns []string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, p := range patterns {
		r.allowedStrayPatterns = append(r.allowedStrayPatterns, compileWildcard(p))
	}
}

func (r *FakeState) AssertSent(f func(client.Request) bool) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, recorded := range r.recorded {
		if f(recorded) {
			return true
		}
	}
	return false
}

func (r *FakeState) AssertSentCount(count int) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.recorded) == count
}

func toHandler(json foundation.Json, v any) func(client.Request) client.Response {
	switch h := v.(type) {
	case func(client.Request) client.Response:
		return h
	case client.Response:
		return func(_ client.Request) client.Response { return h }
	case string:
		return func(_ client.Request) client.Response { return NewResponseFactory(json).String(h, 200) }
	case int:
		return func(_ client.Request) client.Response { return NewResponseFactory(json).Status(h) }
	case *ResponseSequence:
		return func(_ client.Request) client.Response { return h.GetNext() }
	default:
		return func(_ client.Request) client.Response { return NewResponseFactory(json).Status(200) }
	}
}
