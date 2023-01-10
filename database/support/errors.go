package support

import "github.com/pkg/errors"

var (
	ErrorMissingWhereClause = errors.New("WHERE conditions required")
)
