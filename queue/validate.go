package queue

import (
	"errors"
	"reflect"
)

var (
	// ErrTaskMustBeFunc ...
	ErrTaskMustBeFunc = errors.New("handle must be a func type")
	// ErrTaskReturnsNoValue ...
	ErrTaskReturnsNoValue = errors.New("handle must return at least a single value")
	// ErrLastReturnValueMustBeError ..
	ErrLastReturnValueMustBeError = errors.New("last return value of a handle must be error")
)

// ValidateTask validates handle function using reflection and makes sure
// it has a proper signature. Functions used as tasks must return at least a
// single value and the last return type must be error
func ValidateTask(handle any) error {
	v := reflect.ValueOf(handle)
	t := v.Type()

	// Task must be a function
	if t.Kind() != reflect.Func {
		return ErrTaskMustBeFunc
	}

	// Task must return at least a single value
	if t.NumOut() < 1 {
		return ErrTaskReturnsNoValue
	}

	// Last return value must be error
	lastReturnType := t.Out(t.NumOut() - 1)
	errorInterface := reflect.TypeOf((*error)(nil)).Elem()
	if !lastReturnType.Implements(errorInterface) {
		return ErrLastReturnValueMustBeError
	}

	return nil
}
