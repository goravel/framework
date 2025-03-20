package auth

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/cast"

	contractsauth "github.com/goravel/framework/contracts/auth"
	"github.com/goravel/framework/contracts/cache"
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/carbon"
)

type Claims struct {
	Key string `json:"key"`
	jwt.RegisteredClaims
}

const ctxKey = "GoravelAuth"

type Guards map[string]*AuthToken

type AuthToken struct {
	Claims *Claims
	Token  string
}

type JwtGuard struct {
	cache    cache.Cache
	config   config.Config
	ctx      http.Context
	guard    string
	provider contractsauth.UserProvider
}

func NewJwtGuard(guard string, cache cache.Cache, config config.Config, ctx http.Context, provider contractsauth.UserProvider) *JwtGuard {
	return &JwtGuard{
		cache:    cache,
		config:   config,
		ctx:      ctx,
		guard:    guard,
		provider: provider,
	}
}

func (r *JwtGuard) Check() bool {
	if _, err := r.ID(); err != nil {
		return false
	}

	return true
}

func (r *JwtGuard) GetAuthToken() (*AuthToken, error) {
	guards, ok := r.ctx.Value(ctxKey).(Guards)
	if !ok {
		return nil, ErrorParseTokenFirst
	}

	return r.authToken(guards)
}

func (r *JwtGuard) Guest() bool {
	return !r.Check()
}

func (r *JwtGuard) ID() (string, error) {
	guard, err := r.GetAuthToken()
	if err != nil {
		return "", err
	}

	return guard.Claims.Key, nil
}

func (r *JwtGuard) Login(user any) (token string, err error) {
	id := r.provider.GetID(user)
	if id == nil {
		return "", errors.AuthNoPrimaryKeyField
	}

	return r.LoginUsingID(id)
}

func (r *JwtGuard) LoginUsingID(id any) (token string, err error) {
	jwtSecret := r.config.GetString("jwt.secret")
	if jwtSecret == "" {
		return "", errors.AuthEmptySecret
	}

	nowTime := carbon.Now()
	expireTime := nowTime.AddMinutes(r.getTtl()).StdTime()
	key := cast.ToString(id)
	if key == "" {
		return "", errors.AuthInvalidKey
	}
	claims := Claims{
		key,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(nowTime.StdTime()),
			Subject:   r.guard,
		},
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = tokenClaims.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	r.makeAuthContext(&claims, token)

	return
}

func (r *JwtGuard) Logout() error {
	guards, ok := r.ctx.Value(ctxKey).(Guards)
	if !ok {
		return errors.AuthParseTokenFirst
	}

	guard, err := r.authToken(guards)
	if err != nil {
		return err
	}

	if r.cache == nil {
		return errors.CacheSupportRequired.SetModule(errors.ModuleAuth)
	}

	if err := r.cache.Put(getDisabledCacheKey(guard.Token),
		true,
		time.Duration(r.getTtl())*time.Minute,
	); err != nil {
		return err
	}

	delete(guards, r.guard)
	r.ctx.WithValue(ctxKey, guards)

	return nil
}

func (r *JwtGuard) Parse(token string) (*contractsauth.Payload, error) {
	token = strings.ReplaceAll(token, "Bearer ", "")
	if r.cache == nil {
		return nil, errors.CacheSupportRequired.SetModule(errors.ModuleAuth)
	}
	if r.tokenIsDisabled(token) {
		return nil, errors.AuthTokenDisabled
	}

	jwtSecret := r.config.GetString("jwt.secret")
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (any, error) {
		return []byte(jwtSecret), nil
	}, jwt.WithTimeFunc(func() time.Time {
		return carbon.Now().StdTime()
	}))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) && tokenClaims != nil {
			claims, ok := tokenClaims.Claims.(*Claims)
			if !ok {
				return nil, errors.AuthInvalidClaims
			}

			r.makeAuthContext(claims, "")

			return &contractsauth.Payload{
				Guard:    claims.Subject,
				Key:      claims.Key,
				ExpireAt: claims.ExpiresAt.Local(),
				IssuedAt: claims.IssuedAt.Local(),
			}, errors.AuthTokenExpired
		}

		return nil, errors.AuthInvalidToken
	}
	if tokenClaims == nil || !tokenClaims.Valid {
		return nil, errors.AuthInvalidToken
	}

	claims, ok := tokenClaims.Claims.(*Claims)
	if !ok {
		return nil, errors.AuthInvalidClaims
	}

	r.makeAuthContext(claims, token)

	return &contractsauth.Payload{
		Guard:    claims.Subject,
		Key:      claims.Key,
		ExpireAt: claims.ExpiresAt.Time,
		IssuedAt: claims.IssuedAt.Time,
	}, nil
}

// Refresh need parse token first.
func (r *JwtGuard) Refresh() (token string, err error) {
	guards, ok := r.ctx.Value(ctxKey).(Guards)
	if !ok || guards[r.guard] == nil {
		return "", errors.AuthParseTokenFirst
	}
	if guards[r.guard].Claims == nil {
		return "", errors.AuthParseTokenFirst
	}

	nowTime := carbon.Now()
	refreshTtl := r.config.GetInt("jwt.refresh_ttl")
	if refreshTtl == 0 {
		// 100 years
		refreshTtl = 60 * 24 * 365 * 100
	}

	expireTime := carbon.FromStdTime(guards[r.guard].Claims.ExpiresAt.Time).AddMinutes(refreshTtl)
	if nowTime.Gt(expireTime) {
		return "", errors.AuthRefreshTimeExceeded
	}

	return r.LoginUsingID(guards[r.guard].Claims.Key)
}

// User need parse token first.
func (r *JwtGuard) User(user any) error {
	guard, err := r.GetAuthToken()

	if err != nil {
		return err
	}

	err = r.provider.RetriveByID(user, guard.Claims.Key)

	return err
}

func (r *JwtGuard) authToken(guards Guards) (*AuthToken, error) {
	guard, ok := guards[r.guard]
	if !ok || guard == nil {
		return nil, ErrorParseTokenFirst
	}

	if guard.Claims == nil {
		return nil, errors.AuthParseTokenFirst
	}

	if guard.Claims.Key == "" {
		return nil, errors.AuthInvalidKey
	}

	if guard.Token == "" {
		return nil, errors.AuthTokenExpired
	}

	return guard, nil
}

func (r *JwtGuard) getTtl() int {
	var ttl int
	guardTtl := r.config.Get(fmt.Sprintf("auth.guards.%s.ttl", r.guard))
	if guardTtl == nil {
		ttl = r.config.GetInt("jwt.ttl")
	} else {
		ttl = cast.ToInt(guardTtl)
	}

	if ttl == 0 {
		// 100 years
		ttl = 60 * 24 * 365 * 100
	}

	return ttl
}

func (r *JwtGuard) makeAuthContext(claims *Claims, token string) {
	guards, ok := r.ctx.Value(ctxKey).(Guards)
	if !ok {
		guards = make(Guards)
	}
	if guard, ok := guards[r.guard]; ok {
		guard.Claims = claims
		guard.Token = token
	} else {
		guards[r.guard] = &AuthToken{claims, token}
	}
	r.ctx.WithValue(ctxKey, guards)
}

func (r *JwtGuard) tokenIsDisabled(token string) bool {
	return r.cache.GetBool(getDisabledCacheKey(token), false)
}

func getDisabledCacheKey(token string) string {
	return "jwt:disabled:" + token
}
