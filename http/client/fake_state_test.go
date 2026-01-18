package client

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/http/client"
)

type FakeStateTestSuite struct {
	suite.Suite
}

func TestFakeStateTestSuite(t *testing.T) {
	suite.Run(t, new(FakeStateTestSuite))
}

func (s *FakeStateTestSuite) TestMatchSpecificity() {
	mocks := map[string]any{
		"*":                       "generic_wildcard",
		"api.github.com/*":        "domain_wildcard",
		"api.github.com/users/*":  "path_wildcard",
		"api.github.com/user/*":   "path_wildcard_2",
		"api.github.com/users/1":  "specific_id",
		"api.github.com/users/ab": "specific_id_alpha",
	}

	state := NewFakeState(nil, mocks)

	tests := []struct {
		url      string
		expected string
	}{
		{"https://google.com", "generic_wildcard"},
		{"https://api.github.com/repos", "domain_wildcard"},
		{"https://api.github.com/users/goravel", "path_wildcard"},
		// Should match "api.github.com/users/1" (Most specific)
		{"https://api.github.com/users/1", "specific_id"},
		// Should match "api.github.com/users/ab" (Specific alpha ID)
		{"https://api.github.com/users/ab", "specific_id_alpha"},
	}

	for _, tt := range tests {
		req := s.makeHttpRequest(tt.url)
		handler := state.Match(req, "any_client")

		s.NotNil(handler, "Handler should not be nil for %s", tt.url)
		resp := handler(nil)
		body, err := resp.Body()
		s.NoError(err)
		s.Equal(tt.expected, body, "Match logic failed for %s", tt.url)
	}
}

func (s *FakeStateTestSuite) TestMatchMissing() {
	mocks := map[string]any{
		"github": "ok",
	}
	state := NewFakeState(nil, mocks)

	req := s.makeHttpRequest("https://google.com")
	handler := state.Match(req, "dummy_client")

	s.Nil(handler)
}

func (s *FakeStateTestSuite) TestRecord() {
	state := NewFakeState(nil, nil)
	mockReq := NewRequest(nil, nil, "https://api.github.com", "github")
	state.Record(mockReq)

	found := state.AssertSent(func(r client.Request) bool {
		return r.ClientName() == "github"
	})
	s.True(found)

	notFound := state.AssertSent(func(r client.Request) bool {
		return r.ClientName() == "wrong_name"
	})
	s.False(notFound)
}

func (s *FakeStateTestSuite) TestAssertSentCount() {
	state := NewFakeState(nil, nil)

	s.True(state.AssertSentCount(0))

	state.Record(&Request{})
	s.True(state.AssertSentCount(1))
	s.False(state.AssertSentCount(0))

	state.Record(&Request{})
	s.True(state.AssertSentCount(2))
}

func (s *FakeStateTestSuite) TestStrayRequests() {
	state := NewFakeState(nil, nil)

	s.False(state.ShouldPreventStray("https://google.com"))

	state.PreventStrayRequests()
	s.True(state.ShouldPreventStray("https://google.com"))

	state.AllowStrayRequests([]string{"github.com/*"})
	s.False(state.ShouldPreventStray("https://github.com/api"), "Should allow whitelisted pattern")
	s.True(state.ShouldPreventStray("https://google.com"), "Should still block non-whitelisted pattern")
}

func (s *FakeStateTestSuite) TestMockValueConversions() {
	seq := NewFakeSequence(nil)
	seq.PushString(200, "first_event")
	seq.PushString(500, "second_event")

	singleResp := NewFakeResponse(nil).String(202, "static content")

	mocks := map[string]any{
		"https://api.github.com/zen":    "practicality beats purity", // String -> 200 OK + Body
		"https://api.github.com/404":    404,                         // Int -> Status Code Only
		"https://api.github.com/status": singleResp,                  // Response Object
		"https://api.github.com/events": seq,                         // Sequence Object
		"https://api.github.com/func": func(_ client.Request) client.Response {
			return NewFakeResponse(nil).Status(201)
		},
		"https://api.github.com/empty": struct{}{}, // Unknown Type -> 200 OK + Empty
	}

	state := NewFakeState(nil, mocks)

	tests := []struct {
		url    string
		status int
		body   string
	}{
		{"https://api.github.com/zen", 200, "practicality beats purity"},
		{"https://api.github.com/404", 404, ""},
		{"https://api.github.com/func", 201, ""},
		{"https://api.github.com/empty", 200, ""},
	}

	for _, tt := range tests {
		req := s.makeHttpRequest(tt.url)
		resp := state.Match(req, "any")(nil)

		s.Equal(tt.status, resp.Status(), "Failed status check for %s", tt.url)
		if tt.body != "" {
			body, err := resp.Body()
			s.NoError(err)
			s.Equal(tt.body, body)
		}
	}

	reqResp := s.makeHttpRequest("https://api.github.com/status")
	handlerResp := state.Match(reqResp, "any")

	for i := 0; i < 2; i++ {
		resp := handlerResp(nil)
		body, err := resp.Body()
		s.NoError(err)
		s.Equal(202, resp.Status())
		s.Equal("static content", body, "Failed on iteration %d: Body was drained or altered", i+1)
	}

	reqSeq := s.makeHttpRequest("https://api.github.com/events")
	handlerSeq := state.Match(reqSeq, "any")

	resp1 := handlerSeq(nil)
	body1, err := resp1.Body()
	s.NoError(err)
	s.Equal(200, resp1.Status())
	s.Equal("first_event", body1)

	resp2 := handlerSeq(nil)
	body2, err := resp2.Body()
	s.NoError(err)
	s.Equal(500, resp2.Status())
	s.Equal("second_event", body2)
}

func (s *FakeStateTestSuite) makeHttpRequest(link string) *http.Request {
	u, _ := url.Parse(link)
	return &http.Request{URL: u}
}
