package auth

import (
	"errors"
	"fmt"

	contractsauth "github.com/goravel/framework/contracts/auth"
	"github.com/goravel/framework/contracts/cache"
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
)

const ctxKey = "GoravelAuth"

type AuthManager struct {
	contractsauth.Guard
	cache           cache.Cache
	config          config.Config
	ctx             http.Context
	orm             orm.Orm
	guards          map[string]contractsauth.Guard
	providers       map[string]contractsauth.UserProvider
	customGuards    map[string]contractsauth.AuthGuardFunc
	customProviders map[string]contractsauth.UserProviderFunc
}

type Guards map[string]interface{}

func NewAuth(guard string, cache cache.Cache, config config.Config, ctx http.Context, orm orm.Orm) *AuthManager {
	manager := &AuthManager{
		cache:  cache,
		config: config,
		ctx:    ctx,
		orm:    orm,
	}

	guardname := config.GetString("auth.defaults.guard")
	manager.Guard, _ = manager.GetGuard(guardname)

	return manager
}

func (a *AuthManager) Resolve(name string) (contractsauth.Guard, error) {
	driverName := a.config.GetString(fmt.Sprintf("auth.guards.%s.driver", name))
	userProviderName := a.config.GetString(fmt.Sprintf("auth.guards.%s.provider", name))
	provider, err := a.createUserProvider(userProviderName)

	if guardFunc, ok := a.customGuards[driverName]; ok {

		if err != nil {
			return nil, err
		}

		a.guards[name] = guardFunc(name, a, provider)

		return a.guards[name], nil
	}

	switch name {
	case "jwt":
		a.guards[name] = NewJwtGuard(name, a.cache, a.config, a.ctx, provider)
		return a.guards[name], nil
	default:
		return nil, errors.New(fmt.Sprintf("Guard `%s` was not found", name))
	}
}

func (a *AuthManager) createUserProvider(name string) (contractsauth.UserProvider, error) {
	driverName := a.config.GetString(fmt.Sprintf("auth.providers.%s.driver", name))

	if provider, ok := a.providers[name]; ok {
		return provider, nil
	}

	if providerFunc, ok := a.customProviders[driverName]; ok {
		return providerFunc(a), nil
	}

	switch driverName {
	case "orm":
		provider, err := NewOrmUserProvider(driverName, a.orm, a.config)

		if err != nil {
			return nil, err
		}

		a.providers[driverName] = provider
		return a.providers[driverName], nil
	default:
		return nil, errors.New(fmt.Sprintf("User Provider %s was not found", driverName))
	}
}

func (a *AuthManager) GetGuard(name string) (contractsauth.Guard, error) {
	if guard, ok := a.guards[name]; ok {
		return guard, nil
	}
	return a.Resolve(name)
}
