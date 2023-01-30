package auth

import (
	"errors"
)

var (
	ErrorRefreshTimeExceeded = errors.New("refresh time exceeded")
	ErrorTokenExpired        = errors.New("token expired")
	ErrorNoPrimaryKeyField   = errors.New("the primaryKey field was not found in the model, set primaryKey like orm.Model")
	ErrorEmptySecret         = errors.New("secret is required")
	ErrorTokenDisabled       = errors.New("token is disabled")
	ErrorParseTokenFirst     = errors.New("parse token first")
	ErrorInvalidClaims       = errors.New("invalid claims")
	ErrorInvalidToken        = errors.New("invalid token")
	ErrorInvalidKey          = errors.New("invalid key")
)
