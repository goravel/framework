package errors

import (
	"fmt"

	"github.com/goravel/framework/contracts/errors"
)

type errorString struct {
	text         string
	location     string
	args         []any
	withLocation bool
}

// New creates a new error with the provided text and optional location
func New(text string, location ...string) errors.Error {
	err := &errorString{
		text:         text,
		withLocation: true,
	}

	if len(location) > 0 {
		err.location = location[0]
	}

	return err
}

func (e *errorString) Args(args ...any) errors.Error {
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

func (e *errorString) Location(location string) errors.Error {
	e.location = location
	return e
}

func (e *errorString) WithLocation(flag bool) errors.Error {
	e.withLocation = flag
	return e
}
