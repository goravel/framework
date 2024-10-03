package errors

import (
	"errors"
	"fmt"

	contractserrors "github.com/goravel/framework/contracts/errors"
)

type errorString struct {
	text         string
	location     string
	args         []any
	withLocation bool
}

// New creates a new error with the provided text and optional location
func New(text string, location ...string) contractserrors.Error {
	err := &errorString{
		text:         text,
		withLocation: true,
	}

	if len(location) > 0 {
		err.location = location[0]
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

	if e.withLocation && e.location != "" {
		formattedText = fmt.Sprintf("[%s] %s", e.location, formattedText)
	}

	return formattedText
}

func (e *errorString) Location(location string) contractserrors.Error {
	e.location = location
	return e
}

func (e *errorString) WithLocation(flag bool) contractserrors.Error {
	e.withLocation = flag
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
