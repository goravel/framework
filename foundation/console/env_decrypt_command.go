package console

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/errors"
)

type EnvDecryptCommand struct {
}

func NewEnvDecryptCommand() *EnvDecryptCommand {
	return &EnvDecryptCommand{}
}

// Signature The name and signature of the console command.
func (r *EnvDecryptCommand) Signature() string {
	return "env:decrypt"
}

// Description The console command description.
func (r *EnvDecryptCommand) Description() string {
	return "Decrypt an environment file"
}

// Extend The console command extend.
func (r *EnvDecryptCommand) Extend() command.Extend {
	return command.Extend{
		Category: "env",
		Flags: []command.Flag{
			&command.StringFlag{
				Name:    "key",
				Aliases: []string{"k"},
				Value:   "",
				Usage:   "Decryption key",
			},
		},
	}
}

// Handle Execute the console command.
func (r *EnvDecryptCommand) Handle(ctx console.Context) error {
	key := ctx.Option("key")
	if key == "" {
		key = os.Getenv("GORAVEL_ENV_ENCRYPTION_KEY")
	}
	if key == "" {
		ctx.Error("A decryption key is required.")
		return nil
	}
	encryptedData, err := os.ReadFile(".env.encrypted")
	if err != nil {
		ctx.Error("Encrypted environment file not found.")
		return nil
	}
	if _, err = os.Stat(".env"); err == nil {
		ok, _ := ctx.Confirm("Environment file already exists, are you sure to overwrite?", console.ConfirmOption{
			Default:     true,
			Affirmative: "Yes",
			Negative:    "No",
		})
		if !ok {
			return nil
		}
	}
	decrypted, err := r.decrypt(encryptedData, []byte(key))
	if err != nil {
		ctx.Error(fmt.Sprintf("Decrypt error: %v", err))
		return nil
	}
	err = os.WriteFile(".env", decrypted, 0644)
	if err != nil {
		ctx.Error(fmt.Sprintf("Writer error: %v", err))
		return nil
	}
	ctx.Success("Encrypted environment successfully decrypted.")
	return nil
}

func (r *EnvDecryptCommand) decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(string(ciphertext))
	if err != nil {
		return nil, err
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, errors.AesCiphertextInvalid
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)
	return r.pkcs7Unpad(plaintext)
}

func (r *EnvDecryptCommand) pkcs7Unpad(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.AesCiphertextInvalid
	}
	padding := int(data[length-1])
	if padding < 1 || length < padding {
		return nil, errors.AesPaddingInvalid
	}
	return data[:length-padding], nil
}
