package facades

import "errors"

var (
	ErrApplicationNotSet = errors.New("application instance not initialized")
)
