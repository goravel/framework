package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"

	"github.com/goravel/framework/facades"
)

type AES struct {
	key []byte
}

// NewAES returns a new AES hasher.
func NewAES() *AES {
	key := facades.Config.GetString("app.key")
	// check key length before using it
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		panic("[crypt] app.key length must be 16, 24 or 32")
	}
	keyBytes := []byte(key)
	return &AES{
		key: keyBytes,
	}
}

// EncryptString encrypts the given string, and returns the iv and ciphertext as base64 encoded strings.
func (b *AES) EncryptString(value string) (string, string) {
	block, err := aes.NewCipher(b.key)
	if err != nil {
		panic(err.Error())
	}

	plaintext := []byte(value)

	iv := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	ciphertext := aesgcm.Seal(nil, iv, plaintext, nil)

	encodedIv := base64.StdEncoding.EncodeToString(iv)
	encodedCiphertext := base64.StdEncoding.EncodeToString(ciphertext)

	return encodedIv, encodedCiphertext
}

// DecryptString decrypts the given iv and ciphertext, and returns the plaintext.
func (b *AES) DecryptString(iv, ciphertext string) string {
	decodeCiphertext, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		panic(err.Error())
	}
	decodeIv, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		panic(err.Error())
	}

	block, err := aes.NewCipher(b.key)
	if err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	plaintext, err := aesgcm.Open(nil, decodeIv, decodeCiphertext, nil)
	if err != nil {
		panic(err.Error())
	}

	return string(plaintext)
}
