package hash

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/goravel/framework/contracts/config"
)

type Bcrypt struct {
	rounds int
}

// NewBcrypt returns a new Bcrypt hasher.
func NewBcrypt(config config.Config) *Bcrypt {
	return &Bcrypt{
		rounds: config.GetInt("hashing.bcrypt.rounds", 10),
	}
}

// Make returns the hashed value of the given string.
func (b *Bcrypt) Make(value string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(value), b.rounds)
	if err != nil {
		return "", err
	}

	return string(hash), nil
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
		return true
	}
	return hashCost != b.rounds
}
