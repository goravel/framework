package grammars

func prefixArray(prefix string, values []string) []string {
	for i, value := range values {
		values[i] = prefix + " " + value
	}

	return values
}
