package utils

import (
	"slices"
	"strings"

	"github.com/goravel/framework/errors"
)

// validOperators is the whitelist of allowed SQL comparison operators.
var validOperators = []string{
	"=", "<>", "!=", "<", ">", "<=", ">=", "like", "not like",
}

func PrepareWhereOperatorAndValue(args ...any) (op any, value any, err error) {
	if len(args) == 0 || len(args) > 2 {
		return nil, nil, errors.DatabaseInvalidArgumentNumber.Args(len(args), "1 or 2")
	}

	if len(args) == 1 {
		op = "="
		value = args[0]
	} else {
		op = args[0]
		value = args[1]
	}

	// Validate the operator to prevent SQL injection
	if opStr, ok := op.(string); ok {
		if !slices.Contains(validOperators, strings.ToLower(strings.TrimSpace(opStr))) {
			return nil, nil, errors.DatabaseInvalidOperator.Args(opStr)
		}
	}

	return
}
