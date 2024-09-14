package session

import "errors"

var (
	ErrDriverNotSet = errors.New("session driver is not set")
	ErrDriverIsNil  = errors.New("session driver cannot be nil")
)
