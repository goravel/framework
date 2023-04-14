package arr

import (
	"errors"
)

var (
	ErrInvalidKeys           = errors.New("keys should be int or []int")
	ErrInvalidKey            = errors.New("key should be greater than or equal to 0")
	ErrArrayRequired         = errors.New("at least one array is required")
	ErrEmptyArrayNotAllowed  = errors.New("empty array is not allowed")
	ErrNoImplementation      = errors.New("no implementation")
	ErrInvalidRequestedItems = errors.New("requested number of items is greater than the available items")
)
