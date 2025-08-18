package errors

import (
	"errors"
	"fmt"

	contractserrors "github.com/goravel/framework/contracts/errors"
)

type errorString struct {
	text   string
	module string
	args   []any
}

// New creates a new error with the provided text and optional module
func New(text string, module ...string) contractserrors.Error {
	err := &errorString{
		text: text,
	}

	if len(module) > 0 {
		err.module = module[0]
	}

	return err
}

func (e *errorString) Args(args ...any) contractserrors.Error {
	e.args = args
	return e
}

func (e *errorString) Error() string {
	formattedText := e.text

	if len(e.args) > 0 {
		formattedText = fmt.Sprintf(e.text, e.args...)
	}

	if e.module != "" {
		formattedText = fmt.Sprintf("[%s] %s", e.module, formattedText)
	}

	return formattedText
}

func (e *errorString) SetModule(module string) contractserrors.Error {
	e.module = module
	return e
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}

func As(err error, target any) bool {
	return errors.As(err, &target)
}

func Unwrap(err error) error {
	return errors.Unwrap(err)
}

func Ignore(_ error) {}
