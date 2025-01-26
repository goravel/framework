package auth

import (
	"errors"
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
)

type JWTDriver struct {
	auth *Auth
}

func NewJWTDriver(config config.Config, cache Cache, ctx http.Context, orm orm.Orm) *JWTDriver {
	return &JWTDriver{
		auth: NewAuth("jwt", cache, config, ctx, orm),
	}
}

func (d *JWTDriver) Login(userID string, data map[string]interface{}) (string, error) {
	if userID == "" {
		return "", errors.New("user ID is required")
	}

	token, err := d.auth.LoginUsingID(userID)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (d *JWTDriver) Logout(sessionID string) error {
	if sessionID == "" {
		return errors.New("session ID is required")
	}

	d.auth.ctx.WithValue(ctxKey, Guards{
		"jwt": &Guard{
			Token: sessionID,
		},
	})

	err := d.auth.Logout()
	if err != nil {
		return err
	}

	return nil
}

func (d *JWTDriver) Authenticate(sessionID string) error {
	if sessionID == "" {
		return errors.New("session ID is required")
	}

	_, err := d.auth.Parse(sessionID)
	if err != nil {
		return err
	}

	return nil
}
