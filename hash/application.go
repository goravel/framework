package hash

import (
	"github.com/goravel/framework/contracts/hash"
	"github.com/goravel/framework/facades"
)

const (
	DriverArgon2id string = "argon2id"
	DriverBcrypt   string = "bcrypt"
)

type Application struct {
}

func NewApplication() hash.Hash {
	driver := facades.Config.GetString("hashing.driver", "argon2id")

	switch driver {
	case DriverBcrypt:
		return NewBcrypt()
	case DriverArgon2id:
		return NewArgon2id()
	}

	// Default set to argon2id
	return NewArgon2id()
}
