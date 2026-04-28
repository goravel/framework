package logger

import (
	"fmt"
	"reflect"
	"unsafe"
)

var defaultExcludeKeys = []string{
	"GoravelAuthJwt",
	"goravel_http_client_name",
	"locale",
	"fallback_locale",
}

func getContextValues(ctx any, out map[any]any) {
	rv := reflect.ValueOf(ctx)
	if !rv.IsValid() || (rv.Kind() == reflect.Ptr && rv.IsNil()) {
		return
	}
	v := reflect.Indirect(rv)
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return
	}
	t := v.Type()

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

func newExcludeSet(user []string) map[string]struct{} {
	s := make(map[string]struct{}, len(defaultExcludeKeys)+len(user))
	for _, k := range defaultExcludeKeys {
		s[k] = struct{}{}
	}
	for _, k := range user {
		s[k] = struct{}{}
	}
	return s
}

func filterContext(values map[any]any, exclude map[string]struct{}) map[string]any {
	if len(values) == 0 {
		return nil
	}
	var out map[string]any
	for k, v := range values {
		var s string
		if value, ok := k.(string); ok {
			s = value
		} else {
			s = fmt.Sprintf("%v", k)
		}

		if _, drop := exclude[s]; drop {
			continue
		}
		if out == nil {
			out = make(map[string]any)
		}
		out[s] = v
	}
	return out
}
