package logger

import (
	"fmt"
	"reflect"
	"slices"
	"unsafe"
)

// Framework-internal context keys filtered from log output by default.
var defaultContextExcludeKeys = map[string]struct{}{
	"GoravelAuthJwt":           {},
	"goravel_http_client_name": {},
	"locale":                   {},
	"fallback_locale":          {},
}

// getContextValues walks ctx via reflection; the standard library exposes
// no way to enumerate a context's key/value chain.
func getContextValues(ctx any, out map[any]any) {
	v := reflect.Indirect(reflect.ValueOf(ctx))
	t := reflect.TypeOf(ctx)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return
	}

	var key, val any
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if !f.CanAddr() {
			continue
		}
		f = reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
		switch t.Field(i).Name {
		case "Context":
			getContextValues(f.Interface(), out)
		case "key":
			key = f.Interface()
		case "val":
			val = f.Interface()
		}
	}
	if key != nil {
		out[key] = val
	}
}

func filterContextValues(values map[any]any, exclude []string) map[string]any {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]any, len(values))
	for k, v := range values {
		s := fmt.Sprintf("%v", k)
		if _, skip := defaultContextExcludeKeys[s]; skip {
			continue
		}
		if slices.Contains(exclude, s) {
			continue
		}
		out[s] = v
	}
	return out
}
