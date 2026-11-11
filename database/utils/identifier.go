package utils

import (
	"regexp"
	"strings"

	"github.com/goravel/framework/errors"
)

// plainIdentifierRegex matches ASCII identifiers only, optionally qualified by a table. A name with other characters
// (non ASCII letters, "$", etc.) is an expression: the where helpers use it as is, the order helpers reject it.
var plainIdentifierRegex = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)*$`)

// jsonSelectorRegex matches a plain column, optionally qualified by a table, followed by one or more JSON
// path segments, e.g. "data->name" or "users.data->nested->name".
var jsonSelectorRegex = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)*(->[A-Za-z0-9_]+)+$`)

// orderDirectionRegex matches the only directions the order methods accept.
var orderDirectionRegex = regexp.MustCompile(`^(ASC|DESC)$`)

// IsJsonSelector reports whether the column is a JSON selector. It is compiled by the connection's
// grammar (CompileJsonSelector), so it is never pasted into the SQL as is.
func IsJsonSelector(column string) bool {
	return jsonSelectorRegex.MatchString(column)
}

// IsPlainIdentifier reports whether the column is a bare identifier, optionally qualified by a table,
// e.g. "group" or "users.group". Functions, JSON selectors, quoted names and anything else are not plain.
func IsPlainIdentifier(column string) bool {
	return plainIdentifierRegex.MatchString(column)
}

// NormalizeOrderDirection returns the direction in upper case. An empty direction means ASC. Only ASC and
// DESC are accepted; NULLS FIRST / NULLS LAST and any other direction must go through the raw order method.
func NormalizeOrderDirection(direction string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(direction))
	if normalized == "" {
		return "ASC", nil
	}

	if !orderDirectionRegex.MatchString(normalized) {
		return "", errors.DatabaseInvalidOrderByDirection.Args(direction)
	}

	return normalized, nil
}
