package facades

import (
	"github.com/goravel/framework/contracts/auth"
	"github.com/goravel/framework/contracts/auth/access"
)

func Auth() auth.Auth {
	return App().MakeAuth()
}

func Gate() access.Gate {
	return App().MakeGate()
}
