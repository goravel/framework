package client

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/goravel/framework/contracts/http/client"
)

type matchStrategy int

const (
	strategyClient matchStrategy = iota
	strategyURL
	strategyScoped
)

type FakeRule struct {
	pattern    string
	strategy   matchStrategy
	clientName string
	regex      *regexp.Regexp
	handler    func(client.Request) client.Response
}

func NewFakeRule(pattern string, handler func(client.Request) client.Response) *FakeRule {
	var (
		strategy   matchStrategy
		clientName string
		regex      *regexp.Regexp
	)

	if pattern == "*" {
		strategy = strategyURL
		regex = compileWildcard(pattern)

		// The hash symbol(client#path) is the definitive marker for Client Scoping.
	} else if idx := strings.Index(pattern, "#"); idx != -1 {
		strategy = strategyScoped
		clientName = pattern[:idx]
		regex = compileWildcard(pattern[idx+1:])

		// If it has dots, slashes, or colons (and no hash), it is a URL/Path.
		// This correctly handles "localhost:8080", "http://...", and "api.stripe.com".
	} else if strings.ContainsAny(pattern, "./:") || strings.HasPrefix(pattern, "http") {
		strategy = strategyURL
		regex = compileWildcard(pattern)

		// Fallback for simple names like "stripe" or "github"
	} else {
		strategy = strategyClient
		clientName = pattern
	}

	return &FakeRule{
		pattern:    pattern,
		strategy:   strategy,
		clientName: clientName,
		regex:      regex,
		handler:    handler,
	}
}

func (r *FakeRule) Matches(req *http.Request, clientName string) bool {
	switch r.strategy {
	case strategyClient:
		return r.clientName == clientName

	case strategyURL:
		return r.regex.MatchString(req.URL.String())

	case strategyScoped:
		if r.clientName == clientName {
			return r.regex.MatchString(req.URL.Path)
		}
	}

	return false
}

func compileWildcard(p string) *regexp.Regexp {
	if p == "*" {
		return regexp.MustCompile(".*")
	}

	quoted := regexp.QuoteMeta(p)

	expr := strings.ReplaceAll(quoted, "\\*", ".*")

	// If the user provided a full URL (starting with http), strictly anchor start/end.
	// If the user provided a domain (api.stripe.com), allow matching the implicit https:// prefix.
	if strings.HasPrefix(p, "http") {
		expr = "^" + expr + "$"
	} else {
		expr = "^(https?://)?" + expr + "$"
	}

	return regexp.MustCompile(expr)
}
