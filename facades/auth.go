package facades

import (
	"github.com/goravel/framework/contracts/auth"
	"github.com/goravel/framework/contracts/auth/access"
	"github.com/goravel/framework/contracts/http"
)

func Auth(ctx ...http.Context) auth.Auth {
	if len(ctx) > 0 {
		return App().MakeAuth(ctx[0])
	}

	return App().MakeAuth(nil)
}

func Gate() access.Gate {
	return App().MakeGate()
}
