package auth

import (
	"fmt"

	contractsauth "github.com/goravel/framework/contracts/auth"
	"github.com/goravel/framework/contracts/cache"
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/errors"
)

type Auth struct {
	contractsauth.GuardDriver
	cache           cache.Cache
	config          config.Config
	ctx             http.Context
	orm             orm.Orm
	guards          map[string]contractsauth.GuardDriver
	providers       map[string]contractsauth.UserProvider
	customGuards    map[string]contractsauth.GuardFunc
	customProviders map[string]contractsauth.UserProviderFunc
}

func NewAuth(cache cache.Cache, config config.Config, ctx http.Context, orm orm.Orm) (*Auth, error) {
	auth := &Auth{
		cache:           cache,
		config:          config,
		ctx:             ctx,
		orm:             orm,
		guards:          map[string]contractsauth.GuardDriver{},
		providers:       map[string]contractsauth.UserProvider{},
		customGuards:    map[string]contractsauth.GuardFunc{},
		customProviders: map[string]contractsauth.UserProviderFunc{},
	}

	defaultGuard, err := auth.Guard(config.GetString("auth.defaults.guard"))
	if err != nil {
		return nil, err
	}

	auth.GuardDriver = defaultGuard
	return auth, nil
}

func (r *Auth) Extend(name string, fn contractsauth.GuardFunc) {
	r.customGuards[name] = fn
}

func (r *Auth) Guard(name string) (contractsauth.GuardDriver, error) {
	if guard, ok := r.guards[name]; ok {
		return guard, nil
	}
	return r.resolve(name)
}

func (r *Auth) Provider(name string, fn contractsauth.UserProviderFunc) {
	r.customProviders[name] = fn
}

func (r *Auth) createUserProvider(name string) (contractsauth.UserProvider, error) {
	if provider, ok := r.providers[name]; ok {
		return provider, nil
	}

	driverName := r.config.GetString(fmt.Sprintf("auth.providers.%s.driver", name))

	if providerFunc, ok := r.customProviders[driverName]; ok {
		return providerFunc(r)
	}

	switch driverName {
	case "orm":
		provider, err := NewOrmUserProvider(name, r.orm, r.config)

		if err != nil {
			return nil, err
		}

		r.providers[driverName] = provider
		return r.providers[driverName], nil
	default:
		return nil, errors.AuthProviderDriverNotFound.Args(driverName, name)
	}
}

func (r *Auth) resolve(name string) (contractsauth.GuardDriver, error) {
	driverName := r.config.GetString(fmt.Sprintf("auth.guards.%s.driver", name))
	userProviderName := r.config.GetString(fmt.Sprintf("auth.guards.%s.provider", name))
	provider, err := r.createUserProvider(userProviderName)

	if err != nil {
		return nil, err
	}

	if guardFunc, ok := r.customGuards[driverName]; ok {
		if err != nil {
			return nil, err
		}

		guard, err := guardFunc(name, r, provider)
		if err != nil {
			return nil, err
		}

		r.guards[name] = guard

		return guard, nil
	}

	switch driverName {
	case "jwt":
		r.guards[name] = NewJwtGuard(name, r.cache, r.config, r.ctx, provider)
		return r.guards[name], nil
	default:
		return nil, errors.AuthGuardDriverNotFound.Args(driverName, name)
	}
}
