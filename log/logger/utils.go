package logger

import (
	"fmt"
	"reflect"
	"unsafe"
)

// Framework-internal context keys filtered from log output by default.
var defaultExcludeNames = []string{
	"GoravelAuthJwt",
	"goravel_http_client_name",
	"locale",
	"fallback_locale",
}

// excludeSet matches keys two ways: byID by Go identity (struct sentinels
// passed as values) and byName by textual label (defaults + string entries).
type excludeSet struct {
	byID   map[any]struct{}
	byName map[string]struct{}
}

func newExcludeSet(user []any) excludeSet {
	s := excludeSet{
		byID:   map[any]struct{}{},
		byName: make(map[string]struct{}, len(defaultExcludeNames)+len(user)),
	}
	for _, n := range defaultExcludeNames {
		s.byName[n] = struct{}{}
	}
	for _, k := range user {
		if name, ok := k.(string); ok {
			s.byName[name] = struct{}{}
		} else {
			s.byID[k] = struct{}{}
		}
	}
	return s
}

func filterContext(values map[any]any, exclude excludeSet) map[string]any {
	if len(values) == 0 {
		return nil
	}
	// Pick short labels first; escalate any that collide to the full path.
	labels := make(map[any]string, len(values))
	seen := make(map[string]int, len(values))
	for k := range values {
		s := shortName(k)
		labels[k] = s
		seen[s]++
	}

	var out map[string]any
	for k, v := range values {
		if _, drop := exclude.byID[k]; drop {
			continue
		}
		name := labels[k]
		if seen[name] > 1 {
			name = qualifiedName(k)
		}
		if _, drop := exclude.byName[name]; drop {
			continue
		}
		if out == nil {
			out = make(map[string]any)
		}
		out[name] = v
	}
	return out
}

// shortName returns the string content for string-kind keys, %T otherwise.
func shortName(k any) string {
	if s, ok := k.(string); ok {
		return s
	}
	if v := reflect.ValueOf(k); v.IsValid() && v.Kind() == reflect.String {
		return v.String()
	}
	return fmt.Sprintf("%T", k)
}

// qualifiedName escalates to import path + type name when shortName collides.
func qualifiedName(k any) string {
	if t := reflect.TypeOf(k); t != nil && t.PkgPath() != "" && t.Name() != "" {
		return t.PkgPath() + "." + t.Name()
	}
	return fmt.Sprintf("%T", k)
}

// getContextValues walks ctx via reflection; the standard library exposes
// no way to enumerate a context's key/value chain.
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
