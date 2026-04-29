package logger

import (
	"fmt"
	"reflect"
	"unsafe"
)

var defaultExcludeNames = []string{
	"GoravelAuthJwt",
	"goravel_http_client_name",
	"locale",
	"fallback_locale",
}

type excludeSet struct {
	byKey  map[any]struct{}
	byName map[string]struct{}
}

func newExcludeSet(user []any) excludeSet {
	s := excludeSet{
		byKey:  map[any]struct{}{},
		byName: make(map[string]struct{}, len(defaultExcludeNames)+len(user)),
	}
	for _, n := range defaultExcludeNames {
		s.byName[n] = struct{}{}
	}
	for _, k := range user {
		if name, ok := k.(string); ok {
			s.byName[name] = struct{}{}
			continue
		}
		if t := reflect.TypeOf(k); t != nil && t.Comparable() {
			s.byKey[k] = struct{}{}
		}
	}
	return s
}

func filterContext(values map[any]any, exclude excludeSet) map[string]any {
	if len(values) == 0 {
		return nil
	}
	labels := make(map[any]string, len(values))
	seen := make(map[string]int)
	for k := range values {
		if _, drop := exclude.byKey[k]; drop {
			continue
		}
		short := shortName(k)
		if _, drop := exclude.byName[short]; drop {
			continue
		}
		if qual := qualifiedName(k); qual != short {
			if _, drop := exclude.byName[qual]; drop {
				continue
			}
		}
		labels[k] = short
		seen[short]++
	}

	var out map[string]any
	for k, s := range labels {
		if seen[s] > 1 {
			s = qualifiedName(k)
		}
		if out == nil {
			out = make(map[string]any)
		}
		out[s] = values[k]
	}
	return out
}

func shortName(k any) string {
	if s, ok := k.(string); ok {
		return s
	}
	if v := reflect.ValueOf(k); v.IsValid() && v.Kind() == reflect.String {
		return v.String()
	}
	return fmt.Sprintf("%T", k)
}

func qualifiedName(k any) string {
	if t := reflect.TypeOf(k); t != nil && t.PkgPath() != "" && t.Name() != "" {
		return t.PkgPath() + "." + t.Name()
	}
	return fmt.Sprintf("%T", k)
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
