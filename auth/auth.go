package auth

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/cast"
	"gorm.io/gorm/clause"

	contractsauth "github.com/goravel/framework/contracts/auth"
	"github.com/goravel/framework/contracts/cache"
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/database"
)

const ctxKey = "GoravelAuth"

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
	cache  cache.Cache
	config config.Config
	guard  string
	orm    orm.Orm
}

func NewAuth(guard string, cache cache.Cache, config config.Config, orm orm.Orm) *Auth {
	return &Auth{
		cache:  cache,
		config: config,
		guard:  guard,
		orm:    orm,
	}
}

func (a *Auth) Guard(name string) contractsauth.Auth {
	return NewAuth(name, a.cache, a.config, a.orm)
}

// User need parse token first.
func (a *Auth) User(ctx http.Context, user any) error {
	auth, ok := ctx.Value(ctxKey).(Guards)
	if !ok || auth[a.guard] == nil {
		return ErrorParseTokenFirst
	}
	if auth[a.guard].Claims == nil {
		return ErrorParseTokenFirst
	}
	if auth[a.guard].Claims.Key == "" {
		return ErrorInvalidKey
	}
	if auth[a.guard].Token == "" {
		return ErrorTokenExpired
	}
	if err := a.orm.Query().FindOrFail(user, clause.Eq{Column: clause.PrimaryColumn, Value: auth[a.guard].Claims.Key}); err != nil {
		return err
	}

	return nil
}

func (a *Auth) Parse(ctx http.Context, token string) (*contractsauth.Payload, error) {
	token = strings.ReplaceAll(token, "Bearer ", "")
	if a.cache == nil {
		return nil, errors.New("cache support is required")
	}
	if a.tokenIsDisabled(token) {
		return nil, ErrorTokenDisabled
	}

	jwtSecret := a.config.GetString("jwt.secret")
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (any, error) {
		return []byte(jwtSecret), nil
	}, jwt.WithTimeFunc(func() time.Time {
		return carbon.Now().ToStdTime()
	}))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) && tokenClaims != nil {
			claims, ok := tokenClaims.Claims.(*Claims)
			if !ok {
				return nil, ErrorInvalidClaims
			}

			a.makeAuthContext(ctx, claims, "")

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

	a.makeAuthContext(ctx, claims, token)

	return &contractsauth.Payload{
		Guard:    claims.Subject,
		Key:      claims.Key,
		ExpireAt: claims.ExpiresAt.Time,
		IssuedAt: claims.IssuedAt.Time,
	}, nil
}

func (a *Auth) Login(ctx http.Context, user any) (token string, err error) {
	id := database.GetID(user)
	if id == nil {
		return "", ErrorNoPrimaryKeyField
	}

	return a.LoginUsingID(ctx, id)
}

func (a *Auth) LoginUsingID(ctx http.Context, id any) (token string, err error) {
	jwtSecret := a.config.GetString("jwt.secret")
	if jwtSecret == "" {
		return "", ErrorEmptySecret
	}

	nowTime := carbon.Now()
	ttl := a.config.GetInt("jwt.ttl")
	if ttl == 0 {
		// 100 years
		ttl = 60 * 24 * 365 * 100
	}
	expireTime := nowTime.AddMinutes(ttl).ToStdTime()
	key := cast.ToString(id)
	if key == "" {
		return "", ErrorInvalidKey
	}
	claims := Claims{
		key,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(nowTime.ToStdTime()),
			Subject:   a.guard,
		},
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = tokenClaims.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	a.makeAuthContext(ctx, &claims, token)

	return
}

// Refresh need parse token first.
func (a *Auth) Refresh(ctx http.Context) (token string, err error) {
	auth, ok := ctx.Value(ctxKey).(Guards)
	if !ok || auth[a.guard] == nil {
		return "", ErrorParseTokenFirst
	}
	if auth[a.guard].Claims == nil {
		return "", ErrorParseTokenFirst
	}

	nowTime := carbon.Now()
	refreshTtl := a.config.GetInt("jwt.refresh_ttl")
	if refreshTtl == 0 {
		// 100 years
		refreshTtl = 60 * 24 * 365 * 100
	}

	expireTime := carbon.FromStdTime(auth[a.guard].Claims.ExpiresAt.Time).AddMinutes(refreshTtl)
	if nowTime.Gt(expireTime) {
		return "", ErrorRefreshTimeExceeded
	}

	return a.LoginUsingID(ctx, auth[a.guard].Claims.Key)
}

func (a *Auth) Logout(ctx http.Context) error {
	auth, ok := ctx.Value(ctxKey).(Guards)
	if !ok || auth[a.guard] == nil || auth[a.guard].Token == "" {
		return nil
	}

	if a.cache == nil {
		return errors.New("cache support is required")
	}

	ttl := a.config.GetInt("jwt.ttl")
	if ttl == 0 {
		if ok := a.cache.Forever(getDisabledCacheKey(auth[a.guard].Token), true); !ok {
			return errors.New("cache forever failed")
		}
	} else {
		if err := a.cache.Put(getDisabledCacheKey(auth[a.guard].Token),
			true,
			time.Duration(ttl)*time.Minute,
		); err != nil {
			return err
		}
	}

	delete(auth, a.guard)
	ctx.WithValue(ctxKey, auth)

	return nil
}

func (a *Auth) makeAuthContext(ctx http.Context, claims *Claims, token string) {
	ctx.WithValue(ctxKey, Guards{
		a.guard: {claims, token},
	})
}

func (a *Auth) tokenIsDisabled(token string) bool {
	return a.cache.GetBool(getDisabledCacheKey(token), false)
}

func getDisabledCacheKey(token string) string {
	return "jwt:disabled:" + token
}
