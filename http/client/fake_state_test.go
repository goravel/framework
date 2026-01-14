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
		"*":                      "generic_wildcard",
		"api.github.com/*":       "domain_wildcard",
		"api.github.com/users/*": "path_wildcard",
		"api.github.com/users/1": "specific_id",
		"api.github.com/users/a": "specific_id_alpha",
	}

	state := NewFakeState(nil, mocks)

	tests := []struct {
		url      string
		expected string
	}{
		{"https://google.com", "generic_wildcard"},
		{"https://api.github.com/repos", "domain_wildcard"},
		{"https://api.github.com/users/goravel", "path_wildcard"},
		{"https://api.github.com/users/1", "specific_id"},
		{"https://api.github.com/users/a", "specific_id_alpha"},
	}

	for _, tt := range tests {
		req := s.makeHttpRequest(tt.url)
		handler := state.Match(req, "any_client")

		s.NotNil(handler, "Handler should not be nil for %s", tt.url)
		resp := handler(nil)
		body, err := resp.Body()
		s.NoError(err)
		s.Equal(tt.expected, body, "Match failed for %s", tt.url)
	}
}

func (s *FakeStateTestSuite) TestMatchMissing() {
	mocks := map[string]any{
		"github": "ok",
	}
	state := NewFakeState(nil, mocks)

	req := s.makeHttpRequest("https://google.com")
	handler := state.Match(req, "stripe")

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
		return r.ClientName() == "stripe"
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
	s.False(state.ShouldPreventStray("https://github.com/api"))
	s.True(state.ShouldPreventStray("https://google.com"))
}

func (s *FakeStateTestSuite) TestMockValueConversions() {
	seq := NewFakeSequence(nil)
	seq.PushString("first", 200)
	seq.PushString("second", 500)

	singleResp := NewFakeResponse(nil).String("static content", 202)

	mocks := map[string]any{
		"http://string": "ok body",
		"http://int":    404,
		"http://resp":   singleResp,
		"http://seq":    seq,
		"http://func": func(_ client.Request) client.Response {
			return NewFakeResponse(nil).Status(201)
		},
		"http://unknown": struct{}{},
	}

	state := NewFakeState(nil, mocks)

	tests := []struct {
		url    string
		status int
		body   string
	}{
		{"http://string", 200, "ok body"},
		{"http://int", 404, ""},
		{"http://func", 201, ""},
		{"http://unknown", 200, ""},
	}

	for _, tt := range tests {
		req := s.makeHttpRequest(tt.url)
		resp := state.Match(req, "any")(nil)

		s.Equal(tt.status, resp.Status(), "Failed for %s", tt.url)
		if tt.body != "" {
			body, err := resp.Body()
			s.NoError(err)
			s.Equal(tt.body, body)
		}
	}

	reqResp := s.makeHttpRequest("http://resp")
	handlerResp := state.Match(reqResp, "any")

	for i := 0; i < 2; i++ {
		resp := handlerResp(nil)
		body, err := resp.Body()
		s.NoError(err)
		s.Equal(202, resp.Status())
		s.Equal("static content", body, "Failed on iteration %d: Body was drained", i+1)
	}

	reqSeq := s.makeHttpRequest("http://seq")
	handlerSeq := state.Match(reqSeq, "any")

	resp1 := handlerSeq(nil)
	body1, err := resp1.Body()
	s.NoError(err)
	s.Equal(200, resp1.Status())
	s.Equal("first", body1)

	resp2 := handlerSeq(nil)
	body2, err := resp2.Body()
	s.NoError(err)
	s.Equal(500, resp2.Status())
	s.Equal("second", body2)
}

func (s *FakeStateTestSuite) makeHttpRequest(link string) *http.Request {
	u, _ := url.Parse(link)
	return &http.Request{URL: u}
}
