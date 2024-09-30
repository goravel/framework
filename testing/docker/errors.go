package docker

import "errors"

var (
	ErrArtisanNotSet = errors.New("artisan facade is not initialized")
	ErrConfigNotSet  = errors.New("config facade is not initialized")
)
