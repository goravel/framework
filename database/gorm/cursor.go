package gorm

import (
	"reflect"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"gorm.io/gorm"

	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/str"
)

type CursorImpl struct {
	query *QueryImpl
	row   map[string]any
}

func (c *CursorImpl) Scan(value any) error {
	msConfig := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			ToTimeHookFunc(), ToCarbonHookFunc(), ToDeletedAtHookFunc(),
		),
		Squash: true,
		Result: value,
		MatchName: func(mapKey, fieldName string) bool {
			return str.Case2Camel(mapKey) == fieldName || strings.EqualFold(mapKey, fieldName)
		},
	}

	decoder, err := mapstructure.NewDecoder(msConfig)
	if err != nil {
		return err
	}

	if err := decoder.Decode(c.row); err != nil {
		return err
	}

	for relation, args := range c.query.with {
		if err := c.query.origin.Load(value, relation, args...); err != nil {
			return err
		}
	}

	return nil
}

func ToTimeHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if t != reflect.TypeOf(time.Time{}) {
			return data, nil
		}

		switch f.Kind() {
		case reflect.String:
			return time.Parse(time.RFC3339, data.(string))
		case reflect.Float64:
			return time.Unix(0, int64(data.(float64))*int64(time.Millisecond)), nil
		case reflect.Int64:
			return time.Unix(0, data.(int64)*int64(time.Millisecond)), nil
		default:
			return data, nil
		}
	}
}

func ToCarbonHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f == reflect.TypeOf(time.Time{}) {
			switch t {
			case reflect.TypeOf(carbon.DateTime{}):
				return carbon.NewDateTime(carbon.FromStdTime(data.(time.Time))), nil
			case reflect.TypeOf(carbon.DateTimeMilli{}):
				return carbon.NewDateTimeMilli(carbon.FromStdTime(data.(time.Time))), nil
			case reflect.TypeOf(carbon.DateTimeMicro{}):
				return carbon.NewDateTimeMicro(carbon.FromStdTime(data.(time.Time))), nil
			case reflect.TypeOf(carbon.DateTimeNano{}):
				return carbon.NewDateTimeNano(carbon.FromStdTime(data.(time.Time))), nil
			case reflect.TypeOf(carbon.Date{}):
				return carbon.NewDate(carbon.FromStdTime(data.(time.Time))), nil
			case reflect.TypeOf(carbon.DateMilli{}):
				return carbon.NewDateMilli(carbon.FromStdTime(data.(time.Time))), nil
			case reflect.TypeOf(carbon.DateMicro{}):
				return carbon.NewDateMicro(carbon.FromStdTime(data.(time.Time))), nil
			case reflect.TypeOf(carbon.DateNano{}):
				return carbon.NewDateNano(carbon.FromStdTime(data.(time.Time))), nil
			case reflect.TypeOf(carbon.Timestamp{}):
				return carbon.NewTimestamp(carbon.FromStdTime(data.(time.Time))), nil
			case reflect.TypeOf(carbon.TimestampMilli{}):
				return carbon.NewTimestampMilli(carbon.FromStdTime(data.(time.Time))), nil
			case reflect.TypeOf(carbon.TimestampMicro{}):
				return carbon.NewTimestampMicro(carbon.FromStdTime(data.(time.Time))), nil
			case reflect.TypeOf(carbon.TimestampNano{}):
				return carbon.NewTimestampNano(carbon.FromStdTime(data.(time.Time))), nil
			}
		}
		if f.Kind() == reflect.String {
			switch t {
			case reflect.TypeOf(carbon.DateTime{}):
				return carbon.NewDateTime(carbon.Parse(data.(string))), nil
			case reflect.TypeOf(carbon.DateTimeMilli{}):
				return carbon.NewDateTimeMilli(carbon.Parse(data.(string))), nil
			case reflect.TypeOf(carbon.DateTimeMicro{}):
				return carbon.NewDateTimeMicro(carbon.Parse(data.(string))), nil
			case reflect.TypeOf(carbon.DateTimeNano{}):
				return carbon.NewDateTimeNano(carbon.Parse(data.(string))), nil
			case reflect.TypeOf(carbon.Date{}):
				return carbon.NewDate(carbon.Parse(data.(string))), nil
			case reflect.TypeOf(carbon.DateMilli{}):
				return carbon.NewDateMilli(carbon.Parse(data.(string))), nil
			case reflect.TypeOf(carbon.DateMicro{}):
				return carbon.NewDateMicro(carbon.Parse(data.(string))), nil
			case reflect.TypeOf(carbon.DateNano{}):
				return carbon.NewDateNano(carbon.Parse(data.(string))), nil
			case reflect.TypeOf(carbon.Timestamp{}):
				return carbon.NewTimestamp(carbon.Parse(data.(string))), nil
			case reflect.TypeOf(carbon.TimestampMilli{}):
				return carbon.NewTimestampMilli(carbon.Parse(data.(string))), nil
			case reflect.TypeOf(carbon.TimestampMicro{}):
				return carbon.NewTimestampMicro(carbon.Parse(data.(string))), nil
			case reflect.TypeOf(carbon.TimestampNano{}):
				return carbon.NewTimestampNano(carbon.Parse(data.(string))), nil
			}
		}

		return data, nil
	}
}

func ToDeletedAtHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if t != reflect.TypeOf(gorm.DeletedAt{}) {
			return data, nil
		}

		if f == reflect.TypeOf(time.Time{}) {
			return gorm.DeletedAt{Time: data.(time.Time), Valid: true}, nil
		}

		if f.Kind() == reflect.String {
			return gorm.DeletedAt{Time: carbon.Parse(data.(string)).ToStdTime(), Valid: true}, nil
		}

		return data, nil
	}
}
