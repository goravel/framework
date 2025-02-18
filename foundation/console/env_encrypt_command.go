package console

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"os"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/foundation"
)

type EnvEncryptCommand struct {
	app foundation.Application
}

func NewEnvEncryptCommand(app foundation.Application) *EnvEncryptCommand {
	return &EnvEncryptCommand{
		app: app,
	}
}

// Signature The name and signature of the console command.
func (r *EnvEncryptCommand) Signature() string {
	return "env:encrypt"
}

// Description The console command description.
func (r *EnvEncryptCommand) Description() string {
	return "Encrypt an environment file"
}

// Extend The console command extend.
func (r *EnvEncryptCommand) Extend() command.Extend {
	return command.Extend{
		Category: "env",
		Flags: []command.Flag{
			&command.StringFlag{
				Name:    "key",
				Aliases: []string{"k"},
				Value:   "",
				Usage:   "Encryption key",
			},
		},
	}
}

// Handle Execute the console command.
func (r *EnvEncryptCommand) Handle(ctx console.Context) (err error) {
	key := ctx.Option("key")
	if key == "" {
		key = r.app.MakeConfig().GetString("app.key")
	}
	plaintext, err := os.ReadFile(".env")
	if err != nil {
		ctx.Error("Environment file not found.")
		return
	}
	_, err = os.Stat(".env.encrypted")
	if err == nil {
		ok, _ := ctx.Confirm("Encrypted environment file already exists, are you sure to overwrite it?", console.ConfirmOption{
			Default:     true,
			Affirmative: "Yes",
			Negative:    "No",
		})
		if !ok {
			return
		}
	}
	ciphertext, err := encrypt(plaintext, []byte(key))
	if err != nil {
		panic(err)
	}
	base64Data := base64.StdEncoding.EncodeToString(ciphertext)
	err = os.WriteFile(".env.encrypted", []byte(base64Data), 0644)
	if err != nil {
		panic(err)
	}
	ctx.Success("Environment successfully encrypted.")
	return err
}

func encrypt(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	iv := key[:aes.BlockSize]
	plaintext = pkcs7Pad(plaintext, aes.BlockSize)
	mode := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(plaintext))
	mode.CryptBlocks(ciphertext, plaintext)
	return append(iv, ciphertext...), nil
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}
