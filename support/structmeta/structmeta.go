// Package structmeta is only for internal use in the Goravel framework. It provides
// functionality to introspect Go structs and their methods, extracting metadata
// such as field names, types, tags, and method signatures. This is useful for
// performing reflection-based operations on structs, such as generating
// migrations from model definitions or binding methods to runtime instances.
package structmeta

import (
	"reflect"
)

const maxEmbeddedParseDepth = 2

// WalkStruct inspects "v" (which may be a struct value, a *struct, or even a
// nil *struct) and builds a StructMetadata.  If v is not a struct (or pointer to one),
// it returns an empty StructMetadata.
//
// For fields: it flattens up to two levels of embedded structs. For methods:
//   - It finds every method declared on T (value-receiver) and *T (pointer-receiver).
//   - It then tries to "bind" that method onto the runtime instance (v). If v was a
//     non-addressable struct literal or an interface holding a struct, it will make
//     a temporary *T copy so pointer-receivers still work.
//   - If v is a nil pointer, pointer-receiver bindings will remain invalid.
func WalkStruct(v any) StructMetadata {
	if v == nil {
		return StructMetadata{}
	}

	valV := reflect.ValueOf(v)
	typV := valV.Type()

	var typT reflect.Type
	switch typV.Kind() {
	case reflect.Ptr:
		if typV.Elem().Kind() != reflect.Struct {
			return StructMetadata{}
		}
		typT = typV.Elem()

	case reflect.Struct:
		typT = typV

	default:
		return StructMetadata{}
	}

	structName := typT.String()
	fields := parseFields(typT, 0, nil)
	methods := parseMethods(typT, valV)

	return StructMetadata{
		Name:         structName,
		Fields:       fields,
		Methods:      methods,
		ReflectValue: valV,
	}
}

func parseFields(t reflect.Type, depth int, basePath []int) []FieldMetadata {
	var fields []FieldMetadata
	if t.Kind() != reflect.Struct {
		return fields
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		indexPath := append(append([]int(nil), basePath...), i)

		meta := FieldMetadata{
			Name:        f.Name,
			Type:        f.Type.String(),
			Kind:        f.Type.Kind(),
			ReflectType: f.Type,
			Anonymous:   f.Anonymous,
			Tag:         NewTagMetadata(f.Tag),
			IndexPath:   indexPath,
		}
		fields = append(fields, meta)

		if f.Anonymous && depth < maxEmbeddedParseDepth {
			ft := f.Type
			if ft.Kind() == reflect.Ptr {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct {
				embedded := parseFields(ft, depth+1, indexPath)
				fields = append(fields, embedded...)
			}
		}
	}
	return fields
}

func parseMethods(t reflect.Type, v reflect.Value) []MethodMetadata {
	var methods []MethodMetadata

	bindMethod := func(methodName string, wantPtrReceiver bool) reflect.Value {
		if v.Kind() == reflect.Ptr && v.IsNil() {
			return reflect.Value{}
		}
		if wantPtrReceiver && v.Kind() == reflect.Ptr {
			if m := v.MethodByName(methodName); m.IsValid() {
				return m
			}
		}
		if !wantPtrReceiver {
			if m := v.MethodByName(methodName); m.IsValid() {
				return m
			}
		}
		if wantPtrReceiver && v.Kind() == reflect.Struct && v.CanAddr() {
			addrV := v.Addr()
			if m := addrV.MethodByName(methodName); m.IsValid() {
				return m
			}
		}
		if wantPtrReceiver && v.Kind() == reflect.Struct && !v.CanAddr() {
			tmpPtr := reflect.New(v.Type())
			tmpPtr.Elem().Set(v)
			if m := tmpPtr.MethodByName(methodName); m.IsValid() {
				return m
			}
		}
		return reflect.Value{}
	}

	for i := 0; i < t.NumMethod(); i++ {
		mDef := t.Method(i)
		mType := mDef.Type
		var params []string
		for j := 1; j < mType.NumIn(); j++ {
			params = append(params, mType.In(j).String())
		}
		var returns []string
		for j := 0; j < mType.NumOut(); j++ {
			returns = append(returns, mType.Out(j).String())
		}

		bound := bindMethod(mDef.Name, false /*value receiver*/)
		methods = append(methods, MethodMetadata{
			Name:         mDef.Name,
			Receiver:     mType.In(0).String(),
			Parameters:   params,
			Returns:      returns,
			ReflectValue: bound,
		})
	}

	ptrType := reflect.PointerTo(t)
	for i := 0; i < ptrType.NumMethod(); i++ {
		mDef := ptrType.Method(i)
		mType := mDef.Type

		name := mDef.Name
		already := false
		for _, prev := range methods {
			if prev.Name == name {
				already = true
				break
			}
		}
		if already {
			continue
		}

		var params []string
		for j := 1; j < mType.NumIn(); j++ {
			params = append(params, mType.In(j).String())
		}
		var returns []string
		for j := 0; j < mType.NumOut(); j++ {
			returns = append(returns, mType.Out(j).String())
		}

		bound := bindMethod(mDef.Name, true /*pointer receiver*/)
		methods = append(methods, MethodMetadata{
			Name:         mDef.Name,
			Receiver:     mType.In(0).String(),
			Parameters:   params,
			Returns:      returns,
			ReflectValue: bound,
		})
	}

	return methods
}
