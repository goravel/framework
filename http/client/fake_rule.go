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

		// If it starts with http/https, we treat it as a URL immediately.
		// This prevents "https:" from being confused with a Scoped separator.
	} else if strings.HasPrefix(pattern, "http://") || strings.HasPrefix(pattern, "https://") {
		strategy = strategyURL
		regex = compileWildcard(pattern)

		// We've ruled out http/s, so if we see a colon, it must be "client:path".
	} else if idx := strings.Index(pattern, ":"); idx != -1 {
		strategy = strategyScoped
		clientName = pattern[:idx]
		regex = compileWildcard(pattern[idx+1:])

		// If it has a dot (api.stripe.com) or slash (v1/users), it's a URL pattern.
	} else if strings.ContainsAny(pattern, "./") {
		strategy = strategyURL
		regex = compileWildcard(pattern)
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
