package validation

import (
	"net/url"
	"reflect"

	"github.com/gookit/validate"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cast"

	httpvalidate "github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/support/carbon"
)

func init() {
	validate.Config(func(opt *validate.GlobalOption) {
		opt.StopOnError = false
		opt.SkipOnEmpty = true
		opt.FieldTag = "form"
	})
}

var (
	carbonToTransform = map[carbon.Json]func(carbon carbon.Carbon) carbon.Json{
		&carbon.DateTime{}:       carbon.NewDateTime,
		&carbon.DateTimeMilli{}:  carbon.NewDateTimeMilli,
		&carbon.DateTimeMicro{}:  carbon.NewDateTimeMicro,
		&carbon.DateTimeNano{}:   carbon.NewDateTimeNano,
		&carbon.Date{}:           carbon.NewDate,
		&carbon.DateMilli{}:      carbon.NewDateMilli,
		&carbon.DateMicro{}:      carbon.NewDateMicro,
		&carbon.DateNano{}:       carbon.NewDateNano,
		&carbon.Timestamp{}:      carbon.NewTimestamp,
		&carbon.TimestampMilli{}: carbon.NewTimestampMilli,
		&carbon.TimestampMicro{}: carbon.NewTimestampMicro,
		&carbon.TimestampNano{}:  carbon.NewTimestampNano,
	}
)

type Validator struct {
	instance *validate.Validation
	data     validate.DataFace
}

func NewValidator(instance *validate.Validation, data validate.DataFace) *Validator {
	instance.Validate()

	return &Validator{instance: instance, data: data}
}

func (v *Validator) Bind(ptr any) error {
	// Don't bind if there are errors
	if v.Fails() {
		return nil
	}

	var data any
	if formData, ok := v.data.(*validate.FormData); ok {
		values := make(map[string]any)
		for key, value := range v.data.Src().(url.Values) {
			values[key] = value[0]
		}

		for key, value := range formData.Files {
			values[key] = value
		}

		data = values
	} else {
		data = v.data.Src()
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:    "form",
		Result:     &ptr,
		DecodeHook: v.castValue(),
	})
	if err != nil {
		return err
	}

	if err := decoder.Decode(data); err != nil {
		return err
	}

	return nil
}

func (v *Validator) Errors() httpvalidate.Errors {
	if v.instance.Errors == nil || len(v.instance.Errors) == 0 {
		return nil
	}

	return NewErrors(v.instance.Errors)
}

func (v *Validator) Fails() bool {
	return v.instance.IsFail()
}

func (v *Validator) castValue() mapstructure.DecodeHookFunc {
	return func(from reflect.Value, to reflect.Value) (any, error) {
		var (
			err error

			castedValue = from.Interface()
		)

		switch to.Kind() {
		case reflect.String:
			castedValue = cast.ToString(from.Interface())
		case reflect.Int:
			castedValue, err = cast.ToIntE(from.Interface())
		case reflect.Int8:
			castedValue, err = cast.ToInt8E(from.Interface())
		case reflect.Int16:
			castedValue, err = cast.ToInt16E(from.Interface())
		case reflect.Int32:
			castedValue, err = cast.ToInt32E(from.Interface())
		case reflect.Int64:
			castedValue, err = cast.ToInt64E(from.Interface())
		case reflect.Uint:
			castedValue, err = cast.ToUintE(from.Interface())
		case reflect.Uint8:
			castedValue, err = cast.ToUint8E(from.Interface())
		case reflect.Uint16:
			castedValue, err = cast.ToUint16E(from.Interface())
		case reflect.Uint32:
			castedValue, err = cast.ToUint32E(from.Interface())
		case reflect.Uint64:
			castedValue, err = cast.ToUint64E(from.Interface())
		case reflect.Bool:
			castedValue, err = cast.ToBoolE(from.Interface())
		case reflect.Float32:
			castedValue, err = cast.ToFloat32E(from.Interface())
		case reflect.Float64:
			castedValue, err = cast.ToFloat64E(from.Interface())
		case reflect.Slice, reflect.Array:
			switch to.Type().Elem().Kind() {
			case reflect.String:
				castedValue, err = cast.ToStringSliceE(from.Interface())
			case reflect.Int:
				castedValue, err = cast.ToIntSliceE(from.Interface())
			case reflect.Bool:
				castedValue, err = cast.ToBoolSliceE(from.Interface())
			default:
				castedValue, err = cast.ToSliceE(from.Interface())
			}
		case reflect.Map:
			switch to.Type().Key().Kind() {
			case reflect.String:
				castedValue, err = cast.ToStringMapStringE(from.Interface())
			case reflect.Bool:
				castedValue, err = cast.ToStringMapBoolE(from.Interface())
			case reflect.Int:
				castedValue, err = cast.ToStringMapIntE(from.Interface())
			case reflect.Int64:
				castedValue, err = cast.ToStringMapInt64E(from.Interface())
			default:
				castedValue, err = cast.ToStringMapE(from.Interface())
			}
		case reflect.Struct:
			for carbonType, transfrom := range carbonToTransform {
				if to.Type() == reflect.TypeOf(carbonType).Elem() {
					castedValue = castCarbon(from, transfrom)
				}
			}
		default:
			castedValue = from.Interface()
		}

		// Only return casted value if there was no error
		if err == nil {
			return castedValue, nil
		}

		return from.Interface(), nil
	}
}

func castCarbon(from reflect.Value, transfrom func(carbon carbon.Carbon) carbon.Json) any {
	switch len(cast.ToString(from.Interface())) {
	case 10:
		fromInt64, err := cast.ToInt64E(from.Interface())
		if err != nil {
			return transfrom(carbon.Parse(cast.ToString(from.Interface())))
		}
		if fromInt64 > 0 {
			return transfrom(carbon.FromTimestamp(fromInt64))
		}
	case 13:
		fromInt64, err := cast.ToInt64E(from.Interface())
		if err != nil {
			return transfrom(carbon.ParseByFormat(cast.ToString(from.Interface()), "Y-m-d H"))
		}
		if fromInt64 > 0 {
			return transfrom(carbon.FromTimestampMilli(fromInt64))
		}
	case 16:
		fromInt64, err := cast.ToInt64E(from.Interface())
		if err != nil {
			return transfrom(carbon.ParseByFormat(cast.ToString(from.Interface()), "Y-m-d H:i"))
		}
		if fromInt64 > 0 {
			return transfrom(carbon.FromTimestampMicro(fromInt64))
		}
	case 19:
		fromInt64, err := cast.ToInt64E(from.Interface())
		if err != nil {
			return transfrom(carbon.Parse(cast.ToString(from.Interface())))
		}

		if fromInt64 > 0 {
			return transfrom(carbon.FromTimestampNano(fromInt64))
		}
	default:
		return transfrom(carbon.Parse(cast.ToString(from.Interface())))
	}

	return from.Interface()
}
