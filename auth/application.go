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
	"github.com/spf13/cast"
)

var (
	unit = time.Minute

	ErrorRefreshTimeExceeded = errors.New("refresh time exceeded")
	ErrorTokenExpired        = errors.New("token expired")
	ErrorNoPrimaryKeyField   = errors.New("the primaryKey field was not found in the model, set primaryKey like orm.Model")
	ErrorEmptySecret         = errors.New("secret is required")
	ErrorTokenDisabled       = errors.New("token is disabled")
)

type Claims struct {
	Key string `json:"key"`
	jwt.RegisteredClaims
}

type Application struct {
	guards map[string]auth.Auth
	guard  string
	token  string
	claims *Claims
}

func NewApplication(guard string) auth.Auth {
	return &Application{
		guards: make(map[string]auth.Auth),
		guard:  guard,
	}
}

func (app *Application) Guard(name string) auth.Auth {
	if name == facades.Config.GetString("auth.defaults.guard") {
		return app
	}

	guard := app.guards[name]
	if guard != nil {
		return guard
	}

	newApplication := NewApplication(name)
	app.guards[name] = newApplication

	return newApplication
}

//User need parse token first.
func (app *Application) User(user any) error {
	if app.claims == nil {
		return errors.New("parse token first")
	}
	if app.token == "" {
		return ErrorTokenExpired
	}
	if err := facades.Orm.Query().Find(user, app.claims.Key); err != nil {
		return err
	}

	return nil
}

func (app *Application) Parse(token string) error {
	token = strings.ReplaceAll(token, "Bearer ", "")
	if tokenIsDisabled(token) {
		return ErrorTokenDisabled
	}

	jwtSecret := []byte(facades.Config.GetString("jwt.secret"))
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (any, error) {
		return jwtSecret, nil
	})
	if err != nil {
		if strings.Contains(err.Error(), jwt.ErrTokenExpired.Error()) && tokenClaims != nil {
			claims, ok := tokenClaims.Claims.(*Claims)
			if !ok {
				return errors.New("invalid claims")
			}

			app.claims = claims
			app.token = ""

			return ErrorTokenExpired
		} else {
			return err
		}
	}
	if tokenClaims == nil || !tokenClaims.Valid {
		return errors.New("invalid token")
	}

	claims, ok := tokenClaims.Claims.(*Claims)
	if !ok {
		return errors.New("invalid claims")
	}

	if app.claims != claims {
		app.claims = claims
	}
	if app.token != token {
		app.token = token
	}

	return nil
}

func (app *Application) Login(user any) (token string, err error) {
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

	return "", ErrorNoPrimaryKeyField
}

func (app *Application) LoginUsingID(id any) (token string, err error) {
	jwtSecret := facades.Config.GetString("jwt.secret")
	if jwtSecret == "" {
		return "", ErrorEmptySecret
	}

	nowTime := supporttime.Now()
	ttl := facades.Config.GetInt("jwt.ttl")
	expireTime := nowTime.Add(time.Duration(ttl) * unit)
	claims := Claims{
		cast.ToString(id),
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(nowTime),
			Subject:   app.guard,
		},
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = tokenClaims.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	app.claims = &claims
	app.token = token

	return
}

//Refresh need parse token first.
func (app *Application) Refresh() (token string, err error) {
	if app.claims == nil {
		return "", errors.New("parse token first")
	}

	nowTime := supporttime.Now()
	refreshTtl := facades.Config.GetInt("jwt.refresh_ttl")
	expireTime := app.claims.ExpiresAt.Add(time.Duration(refreshTtl) * unit)
	if nowTime.Unix() > expireTime.Unix() {
		return "", ErrorRefreshTimeExceeded
	}

	return app.LoginUsingID(app.claims.Key)
}

func (app *Application) Logout() error {
	if facades.Cache == nil {
		return errors.New("cache support is required")
	}
	if app.token == "" {
		return nil
	}

	if err := facades.Cache.Put(getDisabledCacheKey(app.token), true, time.Duration(facades.Config.GetInt("jwt.ttl"))*unit); err != nil {
		return err
	}

	app.token = ""
	app.claims = nil

	return nil
}

func tokenIsDisabled(token string) bool {
	return facades.Cache.GetBool(getDisabledCacheKey(token), false)
}

func getDisabledCacheKey(token string) string {
	return "jwt:disabled:" + token
}
