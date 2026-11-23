package http

import (
	"net"
	"regexp"
	"slices"
	"strings"

	"github.com/goravel/framework/errors"
)

var (
	domainLabelRegex = regexp.MustCompile(`^[a-z0-9_]([a-z0-9_-]*[a-z0-9_])?$`)
	domainParamRegex = regexp.MustCompile(`^{([A-Za-z_][A-Za-z0-9_]*)}$`)
)

// Domain is the host a route is limited to, such as "admin.example.com", or a
// pattern whose labels in braces are parameters, such as
// "{account}.example.com".
type Domain struct {
	name    string
	params  []string
	pattern *regexp.Regexp
}

// ParseDomain reads a route domain. The domain is matched ignoring case and a
// trailing dot. It is a host name made of labels separated by dots, and a
// label in braces is a parameter that matches one label of the host. An
// empty domain, a scheme, a port, a path, a wildcard, an empty label, a
// parameter that is not a whole label, and a parameter used twice return
// RouteInvalidDomain.
func ParseDomain(domain string) (*Domain, error) {
	name := strings.TrimSuffix(strings.TrimSpace(domain), ".")
	if name == "" {
		return nil, errors.RouteInvalidDomain.Args(domain)
	}

	parsed := &Domain{}
	labels := strings.Split(name, ".")
	patterns := make([]string, len(labels))
	for i, label := range labels {
		if match := domainParamRegex.FindStringSubmatch(label); match != nil {
			if slices.Contains(parsed.params, match[1]) {
				return nil, errors.RouteInvalidDomain.Args(domain)
			}

			parsed.params = append(parsed.params, match[1])
			patterns[i] = "([^.]+)"
			continue
		}

		label = strings.ToLower(label)
		if !domainLabelRegex.MatchString(label) {
			return nil, errors.RouteInvalidDomain.Args(domain)
		}

		labels[i] = label
		patterns[i] = regexp.QuoteMeta(label)
	}

	parsed.name = strings.Join(labels, ".")
	if len(parsed.params) > 0 {
		parsed.pattern = regexp.MustCompile("^" + strings.Join(patterns, `\.`) + "$")
	}

	return parsed, nil
}

// HasParams reports whether the domain has parameters.
func (r *Domain) HasParams() bool {
	return r.pattern != nil
}

// Match reports whether the host, as the request's Host header carries it,
// belongs to the domain, with the values of the domain's parameters.
func (r *Domain) Match(host string) (map[string]string, bool) {
	host = NormalizeHost(host)
	if r.pattern == nil {
		return nil, host == r.name
	}

	matches := r.pattern.FindStringSubmatch(host)
	if matches == nil {
		return nil, false
	}

	params := make(map[string]string, len(r.params))
	for i, name := range r.params {
		params[name] = matches[i+1]
	}

	return params, true
}

// Params returns the names of the domain's parameters, in order.
func (r *Domain) Params() []string {
	return slices.Clone(r.params)
}

// String returns the domain in lower case, keeping its parameters as they
// were written.
func (r *Domain) String() string {
	return r.name
}

// NormalizeHost returns the host in lower case, without its port and
// trailing dot.
func NormalizeHost(host string) string {
	if parsed, _, err := net.SplitHostPort(host); err == nil {
		host = parsed
	}

	return strings.ToLower(strings.TrimSuffix(host, "."))
}
