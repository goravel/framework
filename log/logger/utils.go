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

// excludeSet holds two parallel matchers: byID for identity (==) comparison
// — used when the user excludes a typed value such as a struct sentinel —
// and byName for label comparison, used for framework defaults and any
// string-form entries.
type excludeSet struct {
	byID   map[any]struct{}
	byName map[string]struct{}
}

// newExcludeSet seeds the framework defaults and routes each user entry to
// the appropriate matcher: strings go into the name set, everything else
// goes into the identity set so users can exclude typed-struct sentinels by
// passing the literal value (e.g. middleware.PanicTest{}).
func newExcludeSet(user []any) excludeSet {
	s := excludeSet{
		byID:   make(map[any]struct{}),
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
		s.byID[k] = struct{}{}
	}
	return s
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

func filterContext(values map[any]any, exclude excludeSet) map[string]any {
	if len(values) == 0 {
		return nil
	}
	// Pre-resolve labels with collision-aware naming: a short %T form for
	// the common case, escalating to the fully qualified import path when
	// two keys would otherwise share the same label.
	labels := make(map[any]string, len(values))
	counts := make(map[string]int, len(values))
	for k := range values {
		s := shortKeyName(k)
		labels[k] = s
		counts[s]++
	}

	var out map[string]any
	for k, v := range values {
		if _, drop := exclude.byID[k]; drop {
			continue
		}
		name := labels[k]
		if counts[name] > 1 {
			name = qualifiedKeyName(k)
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

// shortKeyName returns the underlying string for string-kind keys and the
// short %T form (package.Type) for everything else.
func shortKeyName(k any) string {
	if s, ok := k.(string); ok {
		return s
	}
	rv := reflect.ValueOf(k)
	if rv.IsValid() && rv.Kind() == reflect.String {
		return rv.String()
	}
	return fmt.Sprintf("%T", k)
}

// qualifiedKeyName returns the import-path-qualified type name to disambiguate
// keys whose short name collides with another in the same log entry.
func qualifiedKeyName(k any) string {
	if t := reflect.TypeOf(k); t != nil {
		if pkg := t.PkgPath(); pkg != "" && t.Name() != "" {
			return pkg + "." + t.Name()
		}
	}
	return fmt.Sprintf("%T", k)
}
