package validation

import (
	"reflect"

	"github.com/gookit/validate"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cast"

	httpvalidate "github.com/goravel/framework/contracts/validation"
)

func init() {
	validate.Config(func(opt *validate.GlobalOption) {
		opt.StopOnError = false
		opt.SkipOnEmpty = true
		opt.FieldTag = "form"
	})
}

type Validator struct {
	instance *validate.Validation
}

func NewValidator(instance *validate.Validation) *Validator {
	instance.Validate()

	return &Validator{instance: instance}
}

func (v *Validator) Bind(ptr any) error {
	data := v.instance.SafeData()
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:    "form",
		Result:     &ptr,
		DecodeHook: v.castValue(),
	})
	if err != nil {
		return err
	}

	return decoder.Decode(data)
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
		var castedValue any
		var err error

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
