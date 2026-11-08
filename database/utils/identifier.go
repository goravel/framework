package utils

import (
	"regexp"
	"strings"
)

// plainIdentifierRegex matches ASCII identifiers only, optionally qualified by a table. A name with other characters
// (non ASCII letters, "$", etc.) is an expression and must go through the raw order method instead.
var plainIdentifierRegex = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)*$`)

// orderDirectionRegex matches the only directions the order methods accept.
var orderDirectionRegex = regexp.MustCompile(`^(ASC|DESC)$`)

// IsPlainIdentifier reports whether the column is a bare identifier, optionally qualified by a table,
// e.g. "group" or "users.group". Functions, JSON selectors, quoted names and anything else are not plain.
func IsPlainIdentifier(column string) bool {
	return plainIdentifierRegex.MatchString(column)
}

// NormalizeOrderDirection returns the direction in upper case and reports whether the order methods may use it.
// An empty direction means ASC. Only ASC and DESC are accepted; NULLS FIRST / NULLS LAST and any other
// direction must go through the raw order method.
func NormalizeOrderDirection(direction string) (string, bool) {
	normalized := strings.ToUpper(strings.TrimSpace(direction))
	if normalized == "" {
		return "ASC", true
	}

	if !orderDirectionRegex.MatchString(normalized) {
		return "", false
	}

	return normalized, true
}
