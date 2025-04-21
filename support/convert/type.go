package convert

import (
	"fmt"
	"reflect"

	"github.com/spf13/cast"
)

func ToSlice[T int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64](i any) []T {
	v, _ := ToSliceE[T](i)
	return v
}

func ToSliceE[T int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64](i any) ([]T, error) {
	if i == nil {
		return []T{}, fmt.Errorf("unable to cast %#v of type %T", i, i)
	}

	switch v := i.(type) {
	case []T:
		return v, nil
	}

	kind := reflect.TypeOf(i).Kind()
	switch kind {
	case reflect.Slice, reflect.Array:
		s := reflect.ValueOf(i)
		a := make([]T, s.Len())
		for j := 0; j < s.Len(); j++ {
			switch any(a).(type) {
			case []int8:
				val, err := cast.ToInt8E(s.Index(j).Interface())
				if err != nil {
					return []T{}, fmt.Errorf("unable to cast %#v of type %T to []int8", i, i)
				}
				a[j] = T(val)
			case []int16:
				val, err := cast.ToInt16E(s.Index(j).Interface())
				if err != nil {
					return []T{}, fmt.Errorf("unable to cast %#v of type %T to []int32", i, i)
				}
				a[j] = T(val)
			case []int32:
				val, err := cast.ToInt32E(s.Index(j).Interface())
				if err != nil {
					return []T{}, fmt.Errorf("unable to cast %#v of type %T to []int32", i, i)
				}
				a[j] = T(val)
			case []int64:
				val, err := cast.ToInt64E(s.Index(j).Interface())
				if err != nil {
					return []T{}, fmt.Errorf("unable to cast %#v of type %T to []int64", i, i)
				}
				a[j] = T(val)
			case []uint:
				val, err := cast.ToUintE(s.Index(j).Interface())
				if err != nil {
					return []T{}, fmt.Errorf("unable to cast %#v of type %T to []uint", i, i)
				}
				a[j] = T(val)
			case []uint8:
				val, err := cast.ToUint8E(s.Index(j).Interface())
				if err != nil {
					return []T{}, fmt.Errorf("unable to cast %#v of type %T to []uint8", i, i)
				}
				a[j] = T(val)
			case []uint16:
				val, err := cast.ToUint16E(s.Index(j).Interface())
				if err != nil {
					return []T{}, fmt.Errorf("unable to cast %#v of type %T to []uint16", i, i)
				}
				a[j] = T(val)
			case []uint32:
				val, err := cast.ToUint32E(s.Index(j).Interface())
				if err != nil {
					return []T{}, fmt.Errorf("unable to cast %#v of type %T to []uint32", i, i)
				}
				a[j] = T(val)
			case []uint64:
				val, err := cast.ToUint64E(s.Index(j).Interface())
				if err != nil {
					return []T{}, fmt.Errorf("unable to cast %#v of type %T to []uint64", i, i)
				}
				a[j] = T(val)
			case []float32:
				val, err := cast.ToFloat32E(s.Index(j).Interface())
				if err != nil {
					return []T{}, fmt.Errorf("unable to cast %#v of type %T to []float32", i, i)
				}
				a[j] = T(val)
			case []float64:
				val, err := cast.ToFloat64E(s.Index(j).Interface())
				if err != nil {
					return []T{}, fmt.Errorf("unable to cast %#v of type %T to []float64", i, i)
				}
				a[j] = T(val)
			}
		}
		return a, nil
	default:
		return []T{}, fmt.Errorf("unable to cast %#v of type %T", i, i)
	}
}
