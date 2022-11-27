package facades

import (
	"github.com/goravel/framework/contracts/auth"
	"github.com/goravel/framework/contracts/auth/access"
)

var (
	Auth auth.Auth
	Gate access.Gate
)
