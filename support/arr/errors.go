package arr

import (
	"errors"
)

var (
	ErrInvalidKeys          = errors.New("keys should be int or []int")
	ErrInvalidKey           = errors.New("key should be greater than or equal to 0")
	ErrDefaultValueRequired = errors.New("default value is required")
	ErrArrayRequired        = errors.New("at least one array is required")
	ErrEmptySliceNotAllowed = errors.New("empty slice is not allowed")
	ErrExceedMaxLength      = errors.New("exceed max length of slice")
)
