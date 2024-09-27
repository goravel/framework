package session

import "errors"

var (
	ErrDriverNotSet       = errors.New("session driver is not set")
	ErrConfigFacadeNotSet = errors.New("config facade is not initialized")
	ErrJSONNotSet         = errors.New("JSON parser is not initialized")
)
