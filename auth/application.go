package auth

import (
	"errors"
	"reflect"
	"strings"
	"time"

	contractauth "github.com/goravel/framework/contracts/auth"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	supporttime "github.com/goravel/framework/support/time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/cast"
)

const ctxKey = "GoravelAuth"

var (
	unit = time.Minute

	ErrorRefreshTimeExceeded = errors.New("refresh time exceeded")
	ErrorTokenExpired        = errors.New("token expired")
	ErrorNoPrimaryKeyField   = errors.New("the primaryKey field was not found in the model, set primaryKey like orm.Model")
	ErrorEmptySecret         = errors.New("secret is required")
	ErrorTokenDisabled       = errors.New("token is disabled")
	ErrorParseTokenFirst     = errors.New("parse token first")
	ErrorInvalidClaims       = errors.New("invalid claims")
	ErrorInvalidToken        = errors.New("invalid token")
)

type Claims struct {
	Key string `json:"key"`
	jwt.RegisteredClaims
}

type Guard struct {
	Claims *Claims
	Token  string
}

type Auth map[string]*Guard

type Application struct {
	guard string
}

func NewApplication(guard string) contractauth.Auth {
	return &Application{
		guard: guard,
	}
}

func (app *Application) Guard(name string) contractauth.Auth {
	return NewApplication(name)
}

//User need parse token first.
func (app *Application) User(ctx http.Context, user any) error {
	auth, ok := ctx.Value(ctxKey).(Auth)
	if !ok || auth[app.guard] == nil {
		return ErrorParseTokenFirst
	}
	if auth[app.guard].Claims == nil {
		return ErrorParseTokenFirst
	}
	if auth[app.guard].Token == "" {
		return ErrorTokenExpired
	}
	if err := facades.Orm.Query().Find(user, auth[app.guard].Claims.Key); err != nil {
		return err
	}

	return nil
}

func (app *Application) Parse(ctx http.Context, token string) error {
	token = strings.ReplaceAll(token, "Bearer ", "")
	if tokenIsDisabled(token) {
		return ErrorTokenDisabled
	}

	jwtSecret := facades.Config.GetString("jwt.secret")
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (any, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		if strings.Contains(err.Error(), jwt.ErrTokenExpired.Error()) && tokenClaims != nil {
			claims, ok := tokenClaims.Claims.(*Claims)
			if !ok {
				return ErrorInvalidClaims
			}

			app.makeAuthContext(ctx, claims, "")

			return ErrorTokenExpired
		} else {
			return err
		}
	}
	if tokenClaims == nil || !tokenClaims.Valid {
		return ErrorInvalidToken
	}

	claims, ok := tokenClaims.Claims.(*Claims)
	if !ok {
		return ErrorInvalidClaims
	}

	app.makeAuthContext(ctx, claims, token)

	return nil
}

func (app *Application) Login(ctx http.Context, user any) (token string, err error) {
	t := reflect.TypeOf(user).Elem()
	v := reflect.ValueOf(user).Elem()
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Name == "Model" {
			if v.Field(i).Type().Kind() == reflect.Struct {
				structField := v.Field(i).Type()
				for j := 0; j < structField.NumField(); j++ {
					if structField.Field(j).Tag.Get("gorm") == "primaryKey" {
						return app.LoginUsingID(ctx, v.Field(i).Field(j).Interface())
					}
				}
			}
		}
		if t.Field(i).Tag.Get("gorm") == "primaryKey" {
			return app.LoginUsingID(ctx, v.Field(i).Interface())
		}
	}

	return "", ErrorNoPrimaryKeyField
}

func (app *Application) LoginUsingID(ctx http.Context, id any) (token string, err error) {
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

	app.makeAuthContext(ctx, &claims, token)

	return
}

//Refresh need parse token first.
func (app *Application) Refresh(ctx http.Context) (token string, err error) {
	auth, ok := ctx.Value(ctxKey).(Auth)
	if !ok || auth[app.guard] == nil {
		return "", ErrorParseTokenFirst
	}
	if auth[app.guard].Claims == nil {
		return "", ErrorParseTokenFirst
	}

	nowTime := supporttime.Now()
	refreshTtl := facades.Config.GetInt("jwt.refresh_ttl")
	expireTime := auth[app.guard].Claims.ExpiresAt.Add(time.Duration(refreshTtl) * unit)
	if nowTime.Unix() > expireTime.Unix() {
		return "", ErrorRefreshTimeExceeded
	}

	return app.LoginUsingID(ctx, auth[app.guard].Claims.Key)
}

func (app *Application) Logout(ctx http.Context) error {
	auth, ok := ctx.Value(ctxKey).(Auth)
	if !ok || auth[app.guard] == nil || auth[app.guard].Token == "" {
		return nil
	}

	if facades.Cache == nil {
		return errors.New("cache support is required")
	}

	if err := facades.Cache.Put(getDisabledCacheKey(auth[app.guard].Token),
		true,
		time.Duration(facades.Config.GetInt("jwt.ttl"))*unit,
	); err != nil {
		return err
	}

	delete(auth, app.guard)
	ctx.WithValue(ctxKey, auth)

	return nil
}

func (app *Application) makeAuthContext(ctx http.Context, claims *Claims, token string) {
	ctx.WithValue(ctxKey, Auth{
		app.guard: {claims, token},
	})
}

func tokenIsDisabled(token string) bool {
	return facades.Cache.GetBool(getDisabledCacheKey(token), false)
}

func getDisabledCacheKey(token string) string {
	return "jwt:disabled:" + token
}
