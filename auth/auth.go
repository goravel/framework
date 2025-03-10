package auth

import (
	contractsauth "github.com/goravel/framework/contracts/auth"
	"github.com/goravel/framework/contracts/cache"
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
)

const ctxKey = "GoravelAuth"

type AuthManager struct {
	defaultGuard contractsauth.Guard
	cache        cache.Cache
	config       config.Config
	ctx          http.Context
	orm          orm.Orm
}

func NewAuth(guard string, cache cache.Cache, config config.Config, ctx http.Context, orm orm.Orm) *AuthManager {
	return &AuthManager{
		cache:        cache,
		config:       config,
		ctx:          ctx,
		defaultGuard: NewJwtGuard(guard, cache, config, ctx, orm),
		orm:          orm,
	}
}

func (a *AuthManager) Guard(name string) contractsauth.Guard {
	return NewJwtGuard(name, a.cache, a.config, a.ctx, a.orm)
}
