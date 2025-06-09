package auth

import (
	"fmt"

	"github.com/spf13/cast"

	contractsauth "github.com/goravel/framework/contracts/auth"
	"github.com/goravel/framework/contracts/cache"
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/http"
	contractsession "github.com/goravel/framework/contracts/session"
	"github.com/goravel/framework/errors"
)

type SessionGuard struct {
	cache    cache.Cache
	config   config.Config
	session  contractsession.Session
	ctx      http.Context
	guard    string
	provider contractsauth.UserProvider
}

func NewSessionGuard(ctx http.Context, name string, userProvider contractsauth.UserProvider) (contractsauth.GuardDriver, error) {
	if ctx == nil {
		return nil, errors.InvalidHttpContext.SetModule(errors.ModuleAuth)
	}
	if cacheFacade == nil {
		return nil, errors.CacheFacadeNotSet.SetModule(errors.ModuleAuth)
	}
	session := ctx.Request().Session()
	if session == nil {
		return nil, errors.SessionDriverIsNotSet.SetModule(errors.ModuleAuth)
	}

	return &SessionGuard{
		cache:    cacheFacade,
		config:   configFacade,
		session:  session,
		ctx:      ctx,
		guard:    name,
		provider: userProvider,
	}, nil
}

// Check implements auth.GuardDriver.
func (r *SessionGuard) Check() bool {
	_, err := r.ID()

	if err == nil {
		return true
	}

	return false
}

// Guest implements auth.GuardDriver.
func (r *SessionGuard) Guest() bool {
	return !r.Check()
}

// ID implements auth.GuardDriver.
func (r *SessionGuard) ID() (token string, err error) {
	sessionName := r.getSessionName()
	userId := r.session.Get(sessionName, nil)

	if userId == nil {
		return "", errors.AuthInvalidKey
	}

	if id, ok := userId.(string); ok {
		return id, nil
	}

	return "", errors.AuthInvalidKey
}

// Login implements auth.GuardDriver.
func (r *SessionGuard) Login(user any) (token string, err error) {
	id, err := r.provider.GetID(user)
	if err != nil {
		return "", err
	}

	if id, ok := id.(string); ok {
		sessionName := r.getSessionName()
		r.session.Put(sessionName, id)

		return id, nil
	}

	return "", errors.AuthNoPrimaryKeyField
}

// LoginUsingID implements auth.GuardDriver.
func (r *SessionGuard) LoginUsingID(id any) (token string, err error) {
	sessionName := r.getSessionName()
	key := cast.ToString(id)
	if key == "" {
		return "", errors.AuthInvalidKey
	}
	if err := r.provider.RetriveByID(new(interface{}), id); err != nil {
		return "", err
	}
	r.session.Put(sessionName, id)

	return "", nil
}

// Logout implements auth.GuardDriver.
func (r *SessionGuard) Logout() error {
	sessionName := r.getSessionName()
	r.session.Forget(sessionName)

	return nil
}

// Parse implements auth.GuardDriver.
func (r *SessionGuard) Parse(token string) (*contractsauth.Payload, error) {
	panic("unimplemented")
}

// Refresh implements auth.GuardDriver.
func (r *SessionGuard) Refresh() (token string, err error) {
	panic("unimplemented")
}

// User implements auth.GuardDriver.
func (r *SessionGuard) User(user any) error {
	id, err := r.ID()

	if err != nil {
		return err
	}

	return r.provider.RetriveByID(user, id)
}

func (r *SessionGuard) getSessionName() string {
	return fmt.Sprintf("auth_%s_user_id", r.guard)
}
