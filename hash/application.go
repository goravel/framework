package hash

import (
	"github.com/goravel/framework/contracts/hash"
	"github.com/goravel/framework/facades"
)

const (
	DriverArgon2id string = "argon2id"
	DriverBcrypt   string = "bcrypt"
)

func NewApplication() hash.Hash {
	driver := facades.Config.GetString("hashing.driver", "argon2id")

	switch driver {
	case DriverBcrypt:
		return NewBcrypt()
	case DriverArgon2id:
		return NewArgon2id()
	}

	// Default to Argon2id
	return NewArgon2id()
}
