package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"

	"github.com/gookit/color"

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
		color.Redln("[Crypt] Empty or invalid APP_KEY, please reset it.\nRun command:\ngo run . artisan key:generate")
	}
	keyBytes := []byte(key)
	return &AES{
		key: keyBytes,
	}
}

// EncryptString encrypts the given string, and returns the iv and ciphertext as base64 encoded strings.
func (b *AES) EncryptString(value string) string {
	block, err := aes.NewCipher(b.key)
	if err != nil {
		color.Redln("[Crypt] Encrypt init error: " + err.Error())
	}

	plaintext := []byte(value)

	iv := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		color.Redln("[Crypt] Encrypt random iv error: " + err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		color.Redln("[Crypt] Encrypt init error: " + err.Error())
	}

	ciphertext := aesgcm.Seal(nil, iv, plaintext, nil)

	jsonEncoded, err := json.Marshal(map[string][]byte{
		"iv":    iv,
		"value": ciphertext,
	})
	if err != nil {
		color.Redln("[Crypt] Encrypt encode json error: " + err.Error())
	}

	encodedJson := base64.StdEncoding.EncodeToString(jsonEncoded)

	return encodedJson
}

// DecryptString decrypts the given iv and ciphertext, and returns the plaintext.
func (b *AES) DecryptString(payload string) string {
	decodePayload, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		color.Redln("[Crypt] Decrypt payload error: " + err.Error())
	}

	decodeJson := make(map[string][]byte)
	err = json.Unmarshal(decodePayload, &decodeJson)
	if err != nil {
		color.Redln("[Crypt] Decrypt json payload error: " + err.Error())
	}

	decodeIv := decodeJson["iv"]
	decodeCiphertext := decodeJson["value"]

	block, err := aes.NewCipher(b.key)
	if err != nil {
		color.Redln("[Crypt] Decrypt init error: " + err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		color.Redln("[Crypt] Decrypt init error: " + err.Error())
	}

	plaintext, err := aesgcm.Open(nil, decodeIv, decodeCiphertext, nil)
	if err != nil {
		color.Redln("[Crypt] Decrypt plaintext error: " + err.Error())
	}

	return string(plaintext)
}
