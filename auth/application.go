package auth

import (
	"errors"
	"reflect"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/auth"
	"github.com/goravel/framework/facades"
	supporttime "github.com/goravel/framework/support/time"

	"github.com/golang-jwt/jwt/v4"
)

var (
	unit = time.Minute

	RefreshTimeExceeded = errors.New("refresh time exceeded")
	TokenExpired        = errors.New("token expired")
)

type Guard string

type Claims struct {
	Key interface{} `json:"key"`
	jwt.RegisteredClaims
}

type Application struct {
	guard  Guard
	guards map[Guard]interface{}
	claims map[Guard]*Claims
	tokens map[Guard]string
}

func NewApplication() auth.Auth {
	return &Application{
		guard:  "",
		guards: make(map[Guard]interface{}),
		claims: make(map[Guard]*Claims),
		tokens: make(map[Guard]string),
	}
}

func (app *Application) Guard(name string) auth.Auth {
	app.guard = Guard(name)

	return app
}

//User need parse token first.
func (app *Application) User(user interface{}) error {
	guard := app.getGuard()
	if app.guards[guard] != nil {
		user = app.guards[guard]

		return nil
	}

	if app.claims[app.getGuard()] == nil {
		return errors.New("parse token first")
	}

	if app.tokens[app.getGuard()] == "" {
		return TokenExpired
	}

	if err := facades.Orm.Query().Find(user, app.claims[app.getGuard()].Key); err != nil {
		return err
	}

	app.guards[guard] = user

	return nil
}

func (app *Application) Parse(token string) (expired bool, err error) {
	token = strings.ReplaceAll(token, "Bearer ", "")
	if tokenIsDisabled(token) {
		return false, errors.New("token is disabled")
	}

	jwtSecret := []byte(facades.Config.GetString("jwt.secret"))
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		if strings.Contains(err.Error(), jwt.ErrTokenExpired.Error()) && tokenClaims != nil {
			claims, ok := tokenClaims.Claims.(*Claims)
			if !ok {
				return false, errors.New("invalid claims")
			}

			app.claims[app.getGuard()] = claims
			app.tokens[app.getGuard()] = ""

			return true, nil
		} else {
			return false, err
		}
	}
	if tokenClaims == nil || !tokenClaims.Valid {
		return false, errors.New("invalid token")
	}

	claims, ok := tokenClaims.Claims.(*Claims)
	if !ok {
		return false, errors.New("invalid claims")
	}

	app.claims[app.getGuard()] = claims
	app.tokens[app.getGuard()] = token

	return false, nil
}

func (app *Application) Login(user interface{}) (token string, err error) {
	t := reflect.TypeOf(user).Elem()
	v := reflect.ValueOf(user).Elem()
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Name == "Model" {
			if v.Field(i).Type().Kind() == reflect.Struct {
				structField := v.Field(i).Type()
				for j := 0; j < structField.NumField(); j++ {
					if structField.Field(j).Tag.Get("gorm") == "primaryKey" {
						return app.LoginUsingID(v.Field(i).Field(j).Interface())
					}
				}
			}
		}
		if t.Field(i).Tag.Get("gorm") == "primaryKey" {
			return app.LoginUsingID(v.Field(i).Interface())
		}
	}

	return "", errors.New("the primaryKey field was not found in the model, set primaryKey like orm.Model")
}

func (app *Application) LoginUsingID(id interface{}) (token string, err error) {
	jwtSecret := []byte(facades.Config.GetString("jwt.secret"))
	nowTime := supporttime.Now()
	ttl := facades.Config.GetInt("jwt.ttl")
	expireTime := nowTime.Add(time.Duration(ttl) * unit)

	claims := Claims{
		id,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(nowTime),
			Subject:   string(app.getGuard()),
		},
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = tokenClaims.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	app.claims[app.getGuard()] = &claims
	app.tokens[app.getGuard()] = token

	return
}

//Refresh need parse token first.
func (app *Application) Refresh() (token string, err error) {
	if app.claims[app.getGuard()] == nil {
		return "", errors.New("parse token first")
	}

	nowTime := supporttime.Now()
	refreshTtl := facades.Config.GetInt("jwt.refresh_ttl")
	expireTime := app.claims[app.getGuard()].ExpiresAt.Add(time.Duration(refreshTtl) * unit)
	if nowTime.Unix() > expireTime.Unix() {
		return "", RefreshTimeExceeded
	}

	return app.LoginUsingID(app.claims[app.getGuard()].Key)
}

func (app *Application) Logout() error {
	if facades.Cache == nil {
		return errors.New("cache support is required")
	}
	if app.tokens[app.getGuard()] == "" {
		return nil
	}

	if err := facades.Cache.Put(getDisabledCacheKey(app.tokens[app.getGuard()]), true, time.Duration(facades.Config.GetInt("jwt.ttl"))*unit); err != nil {
		return err
	}

	app.guard = ""
	app.tokens[app.getGuard()] = ""
	app.claims[app.getGuard()] = nil
	app.guards[app.getGuard()] = nil

	return nil
}

func (app *Application) getGuard() Guard {
	guard := app.guard
	if app.guard == "" {
		guard = Guard(facades.Config.GetString("auth.defaults.guard"))
	}

	return guard
}

func tokenIsDisabled(token string) bool {
	return facades.Cache.Get(getDisabledCacheKey(token), false).(bool)
}

func getDisabledCacheKey(token string) string {
	return "jwt:disabled:" + token
}
