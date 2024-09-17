package crypt

import (
	"errors"
)

var (
	ErrConfigNotSet          = errors.New("config must not be nil")
	ErrJsonParserNotSet      = errors.New("JSON parser must not be nil")
	ErrAppKeyNotSetInArtisan = errors.New("APP_KEY is required in artisan environment")
	ErrInvalidAppKeyLength   = errors.New("invalid APP_KEY length")
)
