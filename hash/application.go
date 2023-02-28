package hash

import (
	"github.com/goravel/framework/contracts/hash"
	"github.com/goravel/framework/facades"
)

const (
	DriverBcrypt string = "bcrypt"
)

type Application struct {
}

func NewApplication() hash.Hash {
	driver := facades.Config.GetString("hashing.driver", "argon2id")

	if driver == DriverBcrypt {
		return NewBcrypt()
	}

	// Default set to argon2id
	return NewArgon2id()
}
