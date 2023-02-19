package hash

import (
	"github.com/gookit/color"
	"golang.org/x/crypto/bcrypt"

	"github.com/goravel/framework/facades"
)

type Bcrypt struct {
	rounds int
}

// NewBcrypt returns a new Bcrypt hasher.
func NewBcrypt() *Bcrypt {
	return &Bcrypt{
		rounds: facades.Config.GetInt("hashing.bcrypt.rounds", 10),
	}
}

// Make returns the hashed value of the given string.
func (b *Bcrypt) Make(value string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(value), b.rounds)
	if err != nil {
		color.Redln("[Hash] Bcrypt hashing Error : %s", err.Error())
		return ""
	}

	return string(hash)
}

// Check checks if the given string matches the given hash.
func (b *Bcrypt) Check(value, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(value))
	return err == nil
}

// NeedsRehash checks if the given hash needs to be rehashed.
func (b *Bcrypt) NeedsRehash(hash string) bool {
	hashCost, err := bcrypt.Cost([]byte(hash))

	if err != nil {
		return false
	}
	return hashCost != b.rounds
}
