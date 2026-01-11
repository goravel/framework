package client

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/suite"
)

type FakeRuleTestSuite struct {
	suite.Suite
}

func TestFakeRuleTestSuite(t *testing.T) {
	suite.Run(t, new(FakeRuleTestSuite))
}

func (s *FakeRuleTestSuite) TestMatches_ClientStrategy() {
	rule := NewFakeRule("github", nil)
	s.True(rule.Matches(s.makeRequest("https://any.com"), "github"))
	s.False(rule.Matches(s.makeRequest("https://any.com"), "gitlab"))
}

func (s *FakeRuleTestSuite) TestMatches_URLStrategy() {
	rule := NewFakeRule("api.stripe.com/*", nil)
	s.True(rule.Matches(s.makeRequest("https://api.stripe.com/v1/charges"), "any"))
	s.False(rule.Matches(s.makeRequest("https://google.com"), "any"))

	rule = NewFakeRule("*/users", nil)
	s.True(rule.Matches(s.makeRequest("https://example.com/users"), "any"))

	rule = NewFakeRule("https://google.com", nil)
	s.True(rule.Matches(s.makeRequest("https://google.com"), "any"))
	s.False(rule.Matches(s.makeRequest("https://google.com/news"), "any"))
}

func (s *FakeRuleTestSuite) TestMatches_ScopedStrategy() {
	rule := NewFakeRule("github#/repos/*", nil)
	s.True(rule.Matches(s.makeRequest("https://api.github.com/repos/goravel/framework"), "github"))
	s.False(rule.Matches(s.makeRequest("https://api.github.com/repos/goravel/framework"), "gitlab"))
	s.False(rule.Matches(s.makeRequest("https://api.github.com/user"), "github"))
}

func (s *FakeRuleTestSuite) TestMatches_GlobalWildcard() {
	rule := NewFakeRule("*", nil)

	s.True(rule.Matches(s.makeRequest("https://google.com"), "any"))
	s.True(rule.Matches(s.makeRequest("https://facebook.com"), "other"))
}

func (s *FakeRuleTestSuite) TestCompileWildcard() {
	tests := []struct {
		name        string
		pattern     string
		shouldMatch []string
		shouldFail  []string
	}{
		{
			name:    "Implicit Scheme: Domain only should match http/https",
			pattern: "api.github.com/*",
			shouldMatch: []string{
				"https://api.github.com/users",
				"http://api.github.com/repos",
			},
			shouldFail: []string{
				"https://google.com",
				"https://api.github.com.evil.com", // Validates dot escaping
			},
		},
		{
			name:    "Explicit Scheme: Should strictly match the provided scheme",
			pattern: "https://secure.com/*",
			shouldMatch: []string{
				"https://secure.com/login",
			},
			shouldFail: []string{
				"http://secure.com/login",
			},
		},
		{
			name:    "Path Wildcard: Should match anywhere in the path",
			pattern: "*/users/*",
			shouldMatch: []string{
				"https://example.com/users/1",
				"http://localhost/users/create",
			},
			shouldFail: []string{
				"https://example.com/posts/1",
			},
		},
		{
			name:    "Global Catch-All: Should match absolutely anything",
			pattern: "*",
			shouldMatch: []string{
				"https://google.com",
				"random string",
				"",
			},
			shouldFail: []string{},
		},
		{
			name:    "Dot Escaping Security Check: '.' should not match other chars",
			pattern: "goravel.com",
			shouldMatch: []string{
				"https://goravel.com",
				"http://goravel.com",
			},
			shouldFail: []string{
				"https://goravelocom", // Ensures '.' is treated as literal dot, not regex 'any char'
			},
		},
	}

	for _, tt := range tests {
		regex := compileWildcard(tt.pattern)

		for _, u := range tt.shouldMatch {
			s.True(regex.MatchString(u), "Scenario '%s': Pattern '%s' SHOULD match '%s'", tt.name, tt.pattern, u)
		}

		for _, u := range tt.shouldFail {
			s.False(regex.MatchString(u), "Scenario '%s': Pattern '%s' SHOULD NOT match '%s'", tt.name, tt.pattern, u)
		}
	}
}

func (s *FakeRuleTestSuite) makeRequest(link string) *http.Request {
	u, _ := url.Parse(link)
	return &http.Request{URL: u}
}
