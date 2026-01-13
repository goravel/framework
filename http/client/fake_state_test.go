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
		"*":                      "generic_response",
		"api.github.com/*":       "domain_wildcard",
		"api.github.com/users/*": "endpoint_specific",
	}

	state := NewFakeState(nil, mocks)

	req1 := s.makeHttpRequest("https://api.github.com/users/goravel")
	handler1 := state.Match(req1, "any_client")
	resp1 := handler1(nil)
	body1, err1 := resp1.Body()

	s.NotNil(handler1)
	s.NoError(err1)
	s.Equal("endpoint_specific", body1)

	req2 := s.makeHttpRequest("https://api.github.com/repos/create")
	handler2 := state.Match(req2, "any_client")
	resp2 := handler2(nil)
	body2, err2 := resp2.Body()

	s.NoError(err2)
	s.Equal("domain_wildcard", body2)

	req3 := s.makeHttpRequest("https://google.com")
	handler3 := state.Match(req3, "any_client")
	resp3 := handler3(nil)
	body3, err3 := resp3.Body()

	s.NoError(err3)
	s.Equal("generic_response", body3)
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

func (s *FakeStateTestSuite) TestCount() {
	state := NewFakeState(nil, nil)

	state.Record(&Request{})
	state.Record(&Request{})

	s.True(state.AssertSentCount(2))
	s.False(state.AssertSentCount(1))
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
	seq := NewFakeSequence(NewFakeResponse(nil))
	seq.PushString("first", 200)
	seq.PushString("second", 500)

	mocks := map[string]any{
		"http://string": "ok body",
		"http://int":    404,
		"http://func": func(_ client.Request) client.Response {
			return NewFakeResponse(nil).Status(201)
		},
		"http://resp": NewFakeResponse(nil).Status(202),
		"http://nil":  nil,
		"http://seq":  seq,
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
		{"http://resp", 202, ""},
		{"http://nil", 200, ""},
	}

	for _, tt := range tests {
		req := s.makeHttpRequest(tt.url)
		resp := state.Match(req, "any")(nil)

		s.Equal(tt.status, resp.Status(), "Failed for %s", tt.url)
		if tt.body != "" {
			body, _ := resp.Body()
			s.Equal(tt.body, body)
		}
	}

	req := s.makeHttpRequest("http://seq")
	handler := state.Match(req, "any")

	resp1 := handler(nil)
	body1, _ := resp1.Body()
	s.Equal(200, resp1.Status())
	s.Equal("first", body1)

	resp2 := handler(nil)
	body2, _ := resp2.Body()
	s.Equal(500, resp2.Status())
	s.Equal("second", body2)
}

func (s *FakeStateTestSuite) makeHttpRequest(link string) *http.Request {
	u, _ := url.Parse(link)
	return &http.Request{URL: u}
}
