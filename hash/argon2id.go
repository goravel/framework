package hash

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/goravel/framework/facades"
	"golang.org/x/crypto/argon2"
)

type Argon2id struct {
	// The format of the hash.
	format string
	// The version of argon2id to use.
	version int
	// The time cost parameter.
	time uint32
	// The memory cost parameter.
	memory uint32
	// The threads cost parameter.
	threads uint8
	// The length of the key to generate.
	keyLen uint32
	// The length of the random salt to generate.
	saltLen uint32
}

// NewArgon2id returns a new Argon2id hasher.
func NewArgon2id() *Argon2id {
	return &Argon2id{
		format:  "$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		version: 19,
		time:    uint32(facades.Config.GetInt("hashing.argon2id.time", 4)),
		memory:  uint32(facades.Config.GetInt("hashing.argon2id.memory", 65536)),
		threads: uint8(facades.Config.GetInt("hashing.argon2id.threads", 1)),
		keyLen:  uint32(facades.Config.GetInt("hashing.argon2id.keylen", 32)),
		saltLen: uint32(facades.Config.GetInt("hashing.argon2id.saltlen", 16)),
	}
}

func (a *Argon2id) Make(value string) string {
	salt := make([]byte, a.saltLen)
	if _, err := rand.Read(salt); err != nil {
		panic(err.Error())
	}

	hash := argon2.IDKey([]byte(value), salt, a.time, a.memory, a.threads, a.keyLen)

	return fmt.Sprintf(a.format, a.version, a.memory, a.time, a.threads, base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(hash))
}

func (a *Argon2id) Check(value, hash string) bool {
	hashParts := strings.Split(hash, "$")
	if len(hashParts) != 6 {
		return false
	}

	var version int
	_, err := fmt.Sscanf(hashParts[2], "v=%d", &version)
	if err != nil {
		return false
	}
	if version != argon2.Version {
		return false
	}

	_, err = fmt.Sscanf(hashParts[3], "m=%d,t=%d,p=%d", &a.memory, &a.time, &a.threads)
	if err != nil {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(hashParts[4])
	if err != nil {
		return false
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(hashParts[5])
	if err != nil {
		return false
	}

	hashToCompare := argon2.IDKey([]byte(value), salt, a.time, a.memory, a.threads, uint32(len(decodedHash)))

	return subtle.ConstantTimeCompare(decodedHash, hashToCompare) == 1
}

func (a *Argon2id) NeedsRehash(hash string) bool {
	hashParts := strings.Split(hash, "$")
	if len(hashParts) != 6 {
		return true
	}

	var version int
	_, err := fmt.Sscanf(hashParts[2], "v=%d", &version)
	if err != nil {
		return true
	}
	if version != argon2.Version {
		return true
	}

	var memory, time uint32
	var threads uint8
	_, err = fmt.Sscanf(hashParts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	if err != nil {
		return true
	}

	return memory != a.memory || time != a.time || threads != a.threads
}
