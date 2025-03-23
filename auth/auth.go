package auth

import (
	"fmt"
	"sync"

	contractsauth "github.com/goravel/framework/contracts/auth"
	"github.com/goravel/framework/contracts/cache"
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/errors"
)

var (
	guards    = sync.Map{}
	providers = sync.Map{}
)

type Auth struct {
	contractsauth.GuardDriver
	cache           cache.Cache
	config          config.Config
	ctx             http.Context
	log             log.Log
	orm             orm.Orm
	customGuards    map[string]contractsauth.GuardFunc
	customProviders map[string]contractsauth.UserProviderFunc
}

func NewAuth(ctx http.Context, cache cache.Cache, config config.Config, log log.Log, orm orm.Orm) (*Auth, error) {
	auth := &Auth{
		cache:           cache,
		config:          config,
		ctx:             ctx,
		log:             log,
		orm:             orm,
		customGuards:    map[string]contractsauth.GuardFunc{},
		customProviders: map[string]contractsauth.UserProviderFunc{},
	}

	defaultGuard := auth.Guard(config.GetString("auth.defaults.guard"))

	auth.GuardDriver = defaultGuard
	return auth, nil
}

func (r *Auth) Extend(name string, fn contractsauth.GuardFunc) {
	r.customGuards[name] = fn
}

func (r *Auth) Guard(name string) contractsauth.GuardDriver {
	if guard, ok := guards.Load(name); ok {
		return guard.(contractsauth.GuardDriver)
	}

	guard, err := r.resolve(name)
	if err != nil {
		r.log.Panic(err.Error())
		return nil
	}

	return guard
}

func (r *Auth) Provider(name string, fn contractsauth.UserProviderFunc) {
	r.customProviders[name] = fn
}

func (r *Auth) createUserProvider(name string) (contractsauth.UserProvider, error) {
	if provider, ok := providers.Load(name); ok {
		return provider.(contractsauth.UserProvider), nil
	}

	driverName := r.config.GetString(fmt.Sprintf("auth.providers.%s.driver", name))

	if providerFunc, ok := r.customProviders[driverName]; ok {
		provider, err := providerFunc(r)
		if err != nil {
			return nil, err
		}

		providers.Store(driverName, provider)
		return provider, nil
	}

	switch driverName {
	case "orm":
		provider, err := NewOrmUserProvider(name, r.orm, r.config)

		if err != nil {
			return nil, err
		}

		providers.Store(driverName, provider)
		return provider, nil
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
		guard, err := guardFunc(name, r, provider)
		if err != nil {
			return nil, err
		}

		guards.Store(name, guard)

		return guard, nil
	}

	switch driverName {
	case "jwt":
		guard, err := NewJwtGuard(r.ctx, name, r.cache, r.config, provider)
		if err != nil {
			return nil, err
		}

		guards.Store(name, guard)

		return guard, nil
	default:
		return nil, errors.AuthGuardDriverNotFound.Args(driverName, name)
	}
}
