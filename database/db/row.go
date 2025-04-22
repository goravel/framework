package db

import (
	"reflect"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"gorm.io/gorm"

	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/str"
)

type Row struct {
	row map[string]any
}

func NewRow(row map[string]any) *Row {
	return &Row{row: row}
}

func (r *Row) Scan(value any) error {
	msConfig := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			ToStringHookFunc(), ToTimeHookFunc(), ToCarbonHookFunc(), ToDeletedAtHookFunc(),
		),
		Squash: true,
		Result: value,
		MatchName: func(mapKey, fieldName string) bool {
			return str.Of(mapKey).Studly().String() == fieldName || strings.EqualFold(mapKey, fieldName)
		},
	}

	decoder, err := mapstructure.NewDecoder(msConfig)
	if err != nil {
		return err
	}

	return decoder.Decode(r.row)
}

// ToStringHookFunc is a hook function that converts []uint8 to string.
// Mysql returns []uint8 for String type when scanning the rows.
func ToStringHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if t != reflect.TypeOf("") {
			return data, nil
		}

		dataSlice, ok := data.([]uint8)
		if ok {
			return string(dataSlice), nil
		}

		return data, nil
	}
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
			case reflect.TypeOf(carbon.LayoutType[carbon.DateTime]{}):
				return carbon.NewLayoutType[carbon.DateTime](carbon.FromStdTime(data.(time.Time))), nil
			case reflect.TypeOf(carbon.LayoutType[carbon.DateTimeMilli]{}):
				return carbon.NewLayoutType[carbon.DateTimeMilli](carbon.FromStdTime(data.(time.Time))), nil
			case reflect.TypeOf(carbon.LayoutType[carbon.DateTimeMicro]{}):
				return carbon.NewLayoutType[carbon.DateTimeMicro](carbon.FromStdTime(data.(time.Time))), nil
			case reflect.TypeOf(carbon.LayoutType[carbon.DateTimeNano]{}):
				return carbon.NewLayoutType[carbon.DateTimeNano](carbon.FromStdTime(data.(time.Time))), nil
			case reflect.TypeOf(carbon.LayoutType[carbon.Date]{}):
				return carbon.NewLayoutType[carbon.Date](carbon.FromStdTime(data.(time.Time))), nil
			case reflect.TypeOf(carbon.LayoutType[carbon.DateMilli]{}):
				return carbon.NewLayoutType[carbon.DateMilli](carbon.FromStdTime(data.(time.Time))), nil
			case reflect.TypeOf(carbon.LayoutType[carbon.DateMicro]{}):
				return carbon.NewLayoutType[carbon.DateMicro](carbon.FromStdTime(data.(time.Time))), nil
			case reflect.TypeOf(carbon.LayoutType[carbon.DateNano]{}):
				return carbon.NewLayoutType[carbon.DateNano](carbon.FromStdTime(data.(time.Time))), nil
			case reflect.TypeOf(carbon.TimestampType[carbon.Timestamp]{}):
				return carbon.NewTimestampType[carbon.Timestamp](carbon.FromStdTime(data.(time.Time))), nil
			case reflect.TypeOf(carbon.TimestampType[carbon.TimestampMilli]{}):
				return carbon.NewTimestampType[carbon.TimestampMilli](carbon.FromStdTime(data.(time.Time))), nil
			case reflect.TypeOf(carbon.TimestampType[carbon.TimestampMicro]{}):
				return carbon.NewTimestampType[carbon.TimestampMicro](carbon.FromStdTime(data.(time.Time))), nil
			case reflect.TypeOf(carbon.TimestampType[carbon.TimestampNano]{}):
				return carbon.NewTimestampType[carbon.TimestampNano](carbon.FromStdTime(data.(time.Time))), nil
			}
		}
		if f.Kind() == reflect.String {
			switch t {
			case reflect.TypeOf(carbon.LayoutType[carbon.DateTime]{}):
				return carbon.NewLayoutType[carbon.DateTime](carbon.Parse(data.(string))), nil
			case reflect.TypeOf(carbon.LayoutType[carbon.DateTimeMilli]{}):
				return carbon.NewLayoutType[carbon.DateTimeMilli](carbon.Parse(data.(string))), nil
			case reflect.TypeOf(carbon.LayoutType[carbon.DateTimeMicro]{}):
				return carbon.NewLayoutType[carbon.DateTimeMicro](carbon.Parse(data.(string))), nil
			case reflect.TypeOf(carbon.LayoutType[carbon.DateTimeNano]{}):
				return carbon.NewLayoutType[carbon.DateTimeNano](carbon.Parse(data.(string))), nil
			case reflect.TypeOf(carbon.LayoutType[carbon.Date]{}):
				return carbon.NewLayoutType[carbon.Date](carbon.Parse(data.(string))), nil
			case reflect.TypeOf(carbon.LayoutType[carbon.DateMilli]{}):
				return carbon.NewLayoutType[carbon.DateMilli](carbon.Parse(data.(string))), nil
			case reflect.TypeOf(carbon.LayoutType[carbon.DateMicro]{}):
				return carbon.NewLayoutType[carbon.DateMicro](carbon.Parse(data.(string))), nil
			case reflect.TypeOf(carbon.LayoutType[carbon.DateNano]{}):
				return carbon.NewLayoutType[carbon.DateNano](carbon.Parse(data.(string))), nil
			case reflect.TypeOf(carbon.TimestampType[carbon.Timestamp]{}):
				return carbon.NewTimestampType[carbon.Timestamp](carbon.Parse(data.(string))), nil
			case reflect.TypeOf(carbon.TimestampType[carbon.TimestampMilli]{}):
				return carbon.NewTimestampType[carbon.TimestampMilli](carbon.Parse(data.(string))), nil
			case reflect.TypeOf(carbon.TimestampType[carbon.TimestampMicro]{}):
				return carbon.NewTimestampType[carbon.TimestampMicro](carbon.Parse(data.(string))), nil
			case reflect.TypeOf(carbon.TimestampType[carbon.TimestampNano]{}):
				return carbon.NewTimestampType[carbon.TimestampNano](carbon.Parse(data.(string))), nil
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
			return gorm.DeletedAt{Time: carbon.Parse(data.(string)).StdTime(), Valid: true}, nil
		}

		return data, nil
	}
}
