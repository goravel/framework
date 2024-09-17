package filesystem

import "errors"

var (
	ErrConfigFacadeNotSet  = errors.New("config facade not set")
	ErrStorageFacadeNotSet = errors.New("storage facade not set")
)
