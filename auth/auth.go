package auth

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/cast"
	"gorm.io/gorm/clause"

	contractsauth "github.com/goravel/framework/contracts/auth"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/database"
	supporttime "github.com/goravel/framework/support/time"
)

const ctxKey = "GoravelAuth"

var (
	unit = time.Minute
)

type Claims struct {
	Key string `json:"key"`
	jwt.RegisteredClaims
}

type Guard struct {
	Claims *Claims
	Token  string
}

type Guards map[string]*Guard

type Auth struct {
	guard string
}

func NewAuth(guard string) *Auth {
	jwt.TimeFunc = supporttime.Now

	return &Auth{
		guard: guard,
	}
}

func (app *Auth) Guard(name string) contractsauth.Auth {
	return NewAuth(name)
}

//User need parse token first.
func (app *Auth) User(ctx http.Context, user any) error {
	auth, ok := ctx.Value(ctxKey).(Guards)
	if !ok || auth[app.guard] == nil {
		return ErrorParseTokenFirst
	}
	if auth[app.guard].Claims == nil {
		return ErrorParseTokenFirst
	}
	if auth[app.guard].Claims.Key == "" {
		return ErrorInvalidKey
	}
	if auth[app.guard].Token == "" {
		return ErrorTokenExpired
	}
	if err := facades.Orm.Query().Find(user, clause.Eq{Column: clause.PrimaryColumn, Value: auth[app.guard].Claims.Key}); err != nil {
		return err
	}

	return nil
}

func (app *Auth) Parse(ctx http.Context, token string) (*contractsauth.Payload, error) {
	token = strings.ReplaceAll(token, "Bearer ", "")
	if facades.Cache == nil {
		return nil, errors.New("cache support is required")
	}
	if tokenIsDisabled(token) {
		return nil, ErrorTokenDisabled
	}

	jwtSecret := facades.Config.GetString("jwt.secret")
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (any, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) && tokenClaims != nil {
			claims, ok := tokenClaims.Claims.(*Claims)
			if !ok {
				return nil, ErrorInvalidClaims
			}

			app.makeAuthContext(ctx, claims, "")

			return &contractsauth.Payload{
				Guard:    claims.Subject,
				Key:      claims.Key,
				ExpireAt: claims.ExpiresAt.Local(),
				IssuedAt: claims.IssuedAt.Local(),
			}, ErrorTokenExpired
		}

		return nil, ErrorInvalidToken
	}
	if tokenClaims == nil || !tokenClaims.Valid {
		return nil, ErrorInvalidToken
	}

	claims, ok := tokenClaims.Claims.(*Claims)
	if !ok {
		return nil, ErrorInvalidClaims
	}

	app.makeAuthContext(ctx, claims, token)

	return &contractsauth.Payload{
		Guard:    claims.Subject,
		Key:      claims.Key,
		ExpireAt: claims.ExpiresAt.Time,
		IssuedAt: claims.IssuedAt.Time,
	}, nil
}

func (app *Auth) Login(ctx http.Context, user any) (token string, err error) {
	id := database.GetID(user)
	if id == nil {
		return "", ErrorNoPrimaryKeyField
	}

	return app.LoginUsingID(ctx, id)
}

func (app *Auth) LoginUsingID(ctx http.Context, id any) (token string, err error) {
	jwtSecret := facades.Config.GetString("jwt.secret")
	if jwtSecret == "" {
		return "", ErrorEmptySecret
	}

	nowTime := supporttime.Now()
	ttl := facades.Config.GetInt("jwt.ttl")
	expireTime := nowTime.Add(time.Duration(ttl) * unit)
	key := cast.ToString(id)
	if key == "" {
		return "", ErrorInvalidKey
	}
	claims := Claims{
		key,
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
func (app *Auth) Refresh(ctx http.Context) (token string, err error) {
	auth, ok := ctx.Value(ctxKey).(Guards)
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

func (app *Auth) Logout(ctx http.Context) error {
	auth, ok := ctx.Value(ctxKey).(Guards)
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

func (app *Auth) makeAuthContext(ctx http.Context, claims *Claims, token string) {
	ctx.WithValue(ctxKey, Guards{
		app.guard: {claims, token},
	})
}

func tokenIsDisabled(token string) bool {
	return facades.Cache.GetBool(getDisabledCacheKey(token), false)
}

func getDisabledCacheKey(token string) string {
	return "jwt:disabled:" + token
}
