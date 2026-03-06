package validation

import (
	"context"
	"fmt"
	"html"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/spf13/cast"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	validatecontract "github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/errors"
)

// builtinFilters contains all built-in filter functions.
var builtinFilters = map[string]func(val any) any{
	// String cleaning
	"trim": func(val any) any {
		return strings.TrimSpace(cast.ToString(val))
	},
	"ltrim": func(val any) any {
		return strings.TrimLeft(cast.ToString(val), " \t\n\r")
	},
	"rtrim": func(val any) any {
		return strings.TrimRight(cast.ToString(val), " \t\n\r")
	},

	// Case conversion
	"lower": func(val any) any {
		return strings.ToLower(cast.ToString(val))
	},
	"upper": func(val any) any {
		return strings.ToUpper(cast.ToString(val))
	},
	"title": func(val any) any {
		return cases.Title(language.Und).String(cast.ToString(val))
	},
	"ucfirst": func(val any) any {
		s := cast.ToString(val)
		if len(s) == 0 {
			return s
		}
		runes := []rune(s)
		runes[0] = unicode.ToUpper(runes[0])
		return string(runes)
	},
	"lcfirst": func(val any) any {
		s := cast.ToString(val)
		if len(s) == 0 {
			return s
		}
		runes := []rune(s)
		runes[0] = unicode.ToLower(runes[0])
		return string(runes)
	},

	// Naming style
	"camel": func(val any) any {
		return toCamelCase(cast.ToString(val))
	},
	"snake": func(val any) any {
		return toSnakeCase(cast.ToString(val))
	},

	// Type conversion
	"to_int": func(val any) any {
		return cast.ToInt(val)
	},
	"to_uint": func(val any) any {
		return cast.ToUint(val)
	},
	"to_float": func(val any) any {
		return cast.ToFloat64(val)
	},
	"to_bool": func(val any) any {
		return cast.ToBool(val)
	},
	"to_string": func(val any) any {
		return cast.ToString(val)
	},
	"to_time": func(val any) any {
		return cast.ToTime(val)
	},

	// Encoding
	"escape_html": func(val any) any {
		return html.EscapeString(cast.ToString(val))
	},
	"url_encode": func(val any) any {
		return url.QueryEscape(cast.ToString(val))
	},
	"url_decode": func(val any) any {
		decoded, err := url.QueryUnescape(cast.ToString(val))
		if err != nil {
			return cast.ToString(val)
		}
		return decoded
	},

	// Cleaning
	"strip_tags": func(val any) any {
		return stripHTMLTags(cast.ToString(val))
	},
}

// applyFilters applies filter rules to the DataBag.
// Supports wildcard patterns (e.g., "users.*.name": "trim") that expand
// to concrete data paths based on actual data in the bag.
func applyFilters(ctx context.Context, bag *DataBag, filterRules map[string]any, customFilters []validatecontract.Filter) error {
	if len(filterRules) == 0 {
		return nil
	}

	// Build custom filter map by signature
	customFilterMap := make(map[string]validatecontract.Filter)
	for _, f := range customFilters {
		customFilterMap[f.Signature()] = f
	}

	// Parse filter rules, then expand wildcards
	parsed := make(map[string][]ParsedRule)
	for field, filterVal := range filterRules {
		var pf []ParsedRule
		switch v := filterVal.(type) {
		case string:
			pf = ParseRules(v)
		case []string:
			pf = ParseRuleSlice(v)
		default:
			return errors.ValidationInvalidFilterType.Args(field)
		}
		if len(pf) > 0 {
			parsed[field] = pf
		}
	}

	expanded := expandWildcardFields(parsed, bag.Keys(), false)

	for field, parsedFilters := range expanded {
		val, exists := bag.Get(field)
		if !exists {
			continue
		}

		for _, pf := range parsedFilters {
			// Check builtin filters first
			if builtinFn, ok := builtinFilters[pf.Name]; ok {
				val = builtinFn(val)
				continue
			}

			// Check custom filters
			if customFilter, ok := customFilterMap[pf.Name]; ok {
				result, err := callFilterFunc(ctx, customFilter, val, pf.Parameters)
				if err == nil {
					val = result
				}
			}
		}

		_ = bag.Set(field, val)
	}

	return nil
}

// callFilterFunc calls a custom Filter's Handle() returned function via reflection.
// It handles various function signatures:
// - func(val T) R              — single return value
// - func(val T) (R, error)     — return value + error
// - func(val T, args ...A) R   — with extra arguments
func callFilterFunc(ctx context.Context, filter validatecontract.Filter, val any, params []string) (any, error) {
	fn := filter.Handle(ctx)
	if fn == nil {
		return val, fmt.Errorf("filter %s returned nil handler", filter.Signature())
	}

	fnVal := reflect.ValueOf(fn)
	fnType := fnVal.Type()

	if fnType.Kind() != reflect.Func {
		return val, fmt.Errorf("filter %s Handle() must return a function", filter.Signature())
	}

	if fnType.NumIn() == 0 {
		return val, fmt.Errorf("filter %s function must accept at least one argument", filter.Signature())
	}

	// Build argument list
	args := make([]reflect.Value, 0, fnType.NumIn())

	// First argument is the value being filtered
	firstArgType := fnType.In(0)
	convertedVal, err := convertToType(val, firstArgType)
	if err != nil {
		return val, err
	}
	args = append(args, convertedVal)

	// Handle additional parameters
	isVariadic := fnType.IsVariadic()
	if isVariadic {
		// For variadic functions, pass remaining params as variadic args
		variadicType := fnType.In(fnType.NumIn() - 1).Elem()
		for _, p := range params {
			converted, err := convertToType(p, variadicType)
			if err != nil {
				return val, err
			}
			args = append(args, converted)
		}
	} else {
		// For non-variadic functions with extra params
		for i, p := range params {
			argIdx := i + 1
			if argIdx >= fnType.NumIn() {
				break
			}
			argType := fnType.In(argIdx)
			converted, err := convertToType(p, argType)
			if err != nil {
				return val, err
			}
			args = append(args, converted)
		}
	}

	// Call the function
	results := fnVal.Call(args)

	if len(results) == 0 {
		return val, nil
	}

	// Handle return values
	result := results[0].Interface()

	// Check for error return
	if len(results) == 2 {
		if errVal, ok := results[1].Interface().(error); ok && errVal != nil {
			return val, errVal
		}
	}

	return result, nil
}

// convertToType converts a value to the specified reflect.Type.
func convertToType(val any, targetType reflect.Type) (reflect.Value, error) {
	valReflect := reflect.ValueOf(val)

	if !valReflect.IsValid() {
		return reflect.Zero(targetType), nil
	}

	// If assignable directly
	if valReflect.Type().AssignableTo(targetType) {
		return valReflect, nil
	}

	// If convertible
	if valReflect.Type().ConvertibleTo(targetType) {
		return valReflect.Convert(targetType), nil
	}

	// Try using cast for string-based conversions
	strVal := cast.ToString(val)
	switch targetType.Kind() {
	case reflect.String:
		return reflect.ValueOf(strVal), nil
	case reflect.Int:
		return reflect.ValueOf(cast.ToInt(strVal)), nil
	case reflect.Int64:
		return reflect.ValueOf(cast.ToInt64(strVal)), nil
	case reflect.Float64:
		return reflect.ValueOf(cast.ToFloat64(strVal)), nil
	case reflect.Bool:
		return reflect.ValueOf(cast.ToBool(strVal)), nil
	}

	// Use any type as fallback
	if targetType.Kind() == reflect.Interface {
		return valReflect, nil
	}

	return reflect.Zero(targetType), fmt.Errorf("cannot convert %T to %s", val, targetType)
}

// toCamelCase converts a string to camelCase.
func toCamelCase(s string) string {
	words := splitWords(s)
	if len(words) == 0 {
		return ""
	}

	result := strings.ToLower(words[0])
	for _, w := range words[1:] {
		if len(w) > 0 {
			runes := []rune(strings.ToLower(w))
			runes[0] = unicode.ToUpper(runes[0])
			result += string(runes)
		}
	}

	return result
}

// toSnakeCase converts a string to snake_case.
func toSnakeCase(s string) string {
	words := splitWords(s)
	for i, w := range words {
		words[i] = strings.ToLower(w)
	}
	return strings.Join(words, "_")
}

// splitWords splits a string into words based on separators and case changes.
func splitWords(s string) []string {
	// Replace common separators with spaces
	s = strings.NewReplacer("-", " ", "_", " ").Replace(s)

	// Split on camelCase boundaries
	var words []string
	current := strings.Builder{}

	runes := []rune(s)
	for i, r := range runes {
		if r == ' ' {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
			continue
		}

		if i > 0 && unicode.IsUpper(r) && !unicode.IsUpper(runes[i-1]) {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
		}

		current.WriteRune(r)
	}

	if current.Len() > 0 {
		words = append(words, current.String())
	}

	return words
}

var htmlTagRegex = regexp.MustCompile(`<[^>]*>`)

// stripHTMLTags removes HTML tags from a string.
func stripHTMLTags(s string) string {
	return htmlTagRegex.ReplaceAllString(s, "")
}
