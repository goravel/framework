package grammars

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
