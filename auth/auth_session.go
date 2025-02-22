package auth

import (
	"fmt"
	"time"

	"github.com/spf13/cast"
	"gorm.io/gorm/clause"

	contractsauth "github.com/goravel/framework/contracts/auth"
	"github.com/goravel/framework/contracts/cache"
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/database"
)

const sessionCtxKey = "GoravelSessionAuth"

type Session struct {
	SessionID string
}

type Sessions map[string]*Session

type SessionAuth struct {
	cache   cache.Cache
	config  config.Config
	ctx     http.Context
	session string
	orm     orm.Orm
}

func NewSessionAuth(session string, cache cache.Cache, config config.Config, ctx http.Context, orm orm.Orm) *SessionAuth {
	return &SessionAuth{
		cache:   cache,
		config:  config,
		ctx:     ctx,
		session: session,
		orm:     orm,
	}
}

func (a *SessionAuth) Session(name string) contractsauth.Auth {
	return NewAuth(name, a.cache, a.config, a.ctx, a.orm)
}

func (a *SessionAuth) SessionUser(user any) error {
	auth, ok := a.ctx.Value(sessionCtxKey).(Sessions)
	if !ok || auth[a.session] == nil {
		return errors.AuthParseTokenFirst
	}
	if auth[a.session].SessionID == "" {
		return errors.AuthInvalidKey
	}

	if err := a.orm.Query().FindOrFail(user, clause.Eq{Column: clause.PrimaryColumn, Value: auth[a.session].SessionID}); err != nil {
		return err
	}

	return nil
}

func (a *SessionAuth) SessionID() (string, error) {
	auth, ok := a.ctx.Value(sessionCtxKey).(Sessions)
	if !ok || auth[a.session] == nil {
		return "", errors.AuthParseTokenFirst
	}
	if auth[a.session].SessionID == "" {
		return "", errors.AuthInvalidKey
	}

	return auth[a.session].SessionID, nil
}

func (a *SessionAuth) SessionLogin(user any) (string, error) {
	id := database.GetID(user)
	if id == nil {
		return "", errors.AuthNoPrimaryKeyField
	}

	sessionID := cast.ToString(id)
	if sessionID == "" {
		return "", errors.AuthInvalidKey
	}

	if err := a.cache.Put(getSessionCacheKey(sessionID), true, time.Duration(a.getSessionTtl())*time.Minute); err != nil {
		return "", err
	}

	a.makeSessionAuthContext(sessionID)

	return sessionID, nil
}

func (a *SessionAuth) SessionLogout() error {
	auth, ok := a.ctx.Value(sessionCtxKey).(Sessions)
	if !ok || auth[a.session] == nil || auth[a.session].SessionID == "" {
		return nil
	}

	if err := a.cache.Put(getSessionCacheKey(auth[a.session].SessionID), true, time.Duration(a.getSessionTtl())*time.Minute); err != nil {
		return err
	}

	delete(auth, a.session)
	a.ctx.WithValue(sessionCtxKey, auth)

	return nil
}

func (a *SessionAuth) SessionRefresh() (string, error) {
	auth, ok := a.ctx.Value(sessionCtxKey).(Sessions)
	if !ok || auth[a.session] == nil {
		return "", errors.AuthParseTokenFirst
	}

	if !a.cache.GetBool(getSessionCacheKey(auth[a.session].SessionID), false) {
		return "", errors.AuthTokenExpired
	}

	return auth[a.session].SessionID, nil
}

func (a *SessionAuth) makeSessionAuthContext(sessionID string) {
	sessions, ok := a.ctx.Value(sessionCtxKey).(Sessions)
	if !ok {
		sessions = make(Sessions)
	}
	sessions[a.session] = &Session{SessionID: sessionID}
	a.ctx.WithValue(sessionCtxKey, sessions)
}

func (a *SessionAuth) getSessionTtl() int {
	var ttl int
	SessionTtl := a.config.Get(fmt.Sprintf("auth.Sessions.%s.ttl", a.session))
	if SessionTtl == nil {
		ttl = a.config.GetInt("session.ttl")
	} else {
		ttl = cast.ToInt(SessionTtl)
	}

	if ttl == 0 {
		ttl = 60 * 24 * 30
	}

	return ttl
}

func getSessionCacheKey(sessionID string) string {
	return "session:" + sessionID
}
