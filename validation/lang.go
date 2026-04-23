package validation

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
)

//go:embed lang
var LangFS embed.FS

func init() {
	data, err := fs.ReadFile(LangFS, "lang/en/validation.json")
	if err != nil {
		panic(fmt.Sprintf("validation: failed to read embedded lang/en/validation.json: %v", err))
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		panic(fmt.Sprintf("validation: failed to parse embedded lang/en/validation.json: %v", err))
	}

	defaultMessages = flattenMessages("", raw)
}

// flattenMessages flattens a nested JSON map into dot-separated keys.
// e.g. {"min": {"string": "...", "numeric": "..."}} → {"min.string": "...", "min.numeric": "..."}
func flattenMessages(prefix string, m map[string]any) map[string]string {
	result := make(map[string]string)
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		switch val := v.(type) {
		case string:
			result[key] = val
		case map[string]any:
			for fk, fv := range flattenMessages(key, val) {
				result[fk] = fv
			}
		}
	}
	return result
}
