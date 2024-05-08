package grammars

import (
	"github.com/spf13/cast"
)

func getDefaultValue(def any) string {
	switch def.(type) {
	case bool:
		return "'" + cast.ToString(cast.ToInt(def)) + "'"
	}

	return "'" + cast.ToString(def) + "'"
}

func prefixArray(prefix string, values []string) []string {
	for i, value := range values {
		values[i] = prefix + " " + value
	}

	return values
}

func quoteString(value []string) []string {
	for i, v := range value {
		value[i] = "'" + v + "'"
	}

	return value
}
