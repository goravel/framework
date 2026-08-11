// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0
//
// Adapted from: https://github.com/open-telemetry/opentelemetry-go-contrib
// Modified by Goravel to support framework-specific log contracts.

package log

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"

	contractslog "github.com/goravel/framework/contracts/log"
)

func toSeverity(level contractslog.Level) otellog.Severity {
	switch level {
	case contractslog.LevelPanic:
		return otellog.SeverityFatal4
	case contractslog.LevelFatal:
		return otellog.SeverityFatal
	case contractslog.LevelError:
		return otellog.SeverityError
	case contractslog.LevelWarning:
		return otellog.SeverityWarn
	case contractslog.LevelInfo:
		return otellog.SeverityInfo
	case contractslog.LevelDebug:
		return otellog.SeverityDebug
	default:
		return otellog.SeverityInfo
	}
}

func toValue(v any) attribute.Value {
	if v == nil {
		return attribute.Value{}
	}

	switch val := v.(type) {
	case attribute.Value:
		return val
	case string:
		return attribute.StringValue(val)
	case bool:
		return attribute.BoolValue(val)
	case int:
		return attribute.Int64Value(int64(val))
	case int64:
		return attribute.Int64Value(val)
	case float64:
		return attribute.Float64Value(val)
	case error:
		return attribute.StringValue(val.Error())
	case []byte:
		return attribute.ByteSliceValue(val)
	case time.Time:
		return attribute.StringValue(val.Format(time.RFC3339Nano))
	case map[string]any:
		return toMapValue(val)
	case []string:
		return toStringSliceValue(val)
	case int32:
		return attribute.Int64Value(int64(val))
	case int16:
		return attribute.Int64Value(int64(val))
	case int8:
		return attribute.Int64Value(int64(val))
	case uint:
		return toUintValue(uint64(val))
	case uint64:
		return toUintValue(val)
	case uint32:
		return attribute.Int64Value(int64(val))
	case uint16:
		return attribute.Int64Value(int64(val))
	case uint8:
		return attribute.Int64Value(int64(val))
	case uintptr:
		return toUintValue(uint64(val))
	case float32:
		return attribute.Float64Value(float64(val))
	case time.Duration:
		return attribute.Int64Value(val.Nanoseconds())
	case complex64:
		r := attribute.Float64("r", float64(real(val)))
		i := attribute.Float64("i", float64(imag(val)))
		return attribute.MapValue(r, i)
	case complex128:
		r := attribute.Float64("r", real(val))
		i := attribute.Float64("i", imag(val))
		return attribute.MapValue(r, i)
	case fmt.Stringer:
		return attribute.StringValue(val.String())
	}

	return toReflectedValue(v)
}

func toReflectedValue(v any) attribute.Value {
	t := reflect.TypeOf(v)
	if t == nil {
		return attribute.Value{}
	}
	val := reflect.ValueOf(v)

	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		items := make([]attribute.Value, val.Len())
		for i := 0; i < val.Len(); i++ {
			items[i] = toValue(val.Index(i).Interface())
		}
		return attribute.SliceValue(items...)

	case reflect.Map:
		kvs := make([]attribute.KeyValue, 0, val.Len())
		iter := val.MapRange()
		for iter.Next() {
			k := iter.Key()
			var keyStr string
			if k.Kind() == reflect.String {
				keyStr = k.String()
			} else {
				keyStr = fmt.Sprintf("%+v", k.Interface())
			}
			kvs = append(kvs, attribute.KeyValue{
				Key:   attribute.Key(keyStr),
				Value: toValue(iter.Value().Interface()),
			})
		}
		return attribute.MapValue(kvs...)

	case reflect.Struct:
		return attribute.StringValue(fmt.Sprintf("%+v", v))

	case reflect.Pointer, reflect.Interface:
		if val.IsNil() {
			return attribute.Value{}
		}
		return toValue(val.Elem().Interface())

	case reflect.String:
		return attribute.StringValue(val.String())
	case reflect.Bool:
		return attribute.BoolValue(val.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return attribute.Int64Value(val.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return toUintValue(val.Uint())
	case reflect.Float32, reflect.Float64:
		return attribute.Float64Value(val.Float())

	default:
		return attribute.StringValue(fmt.Sprintf("unhandled: (%s) %+v", t, v))
	}
}

func toMapValue(m map[string]any) attribute.Value {
	kvs := make([]attribute.KeyValue, 0, len(m))
	for k, v := range m {
		kvs = append(kvs, attribute.KeyValue{
			Key:   attribute.Key(k),
			Value: toValue(v),
		})
	}
	return attribute.MapValue(kvs...)
}

func toStringSliceValue(s []string) attribute.Value {
	items := make([]attribute.Value, len(s))
	for i, v := range s {
		items[i] = attribute.StringValue(v)
	}
	return attribute.SliceValue(items...)
}

func toUintValue(v uint64) attribute.Value {
	if v > math.MaxInt64 {
		return attribute.StringValue(strconv.FormatUint(v, 10))
	}
	return attribute.Int64Value(int64(v))
}
