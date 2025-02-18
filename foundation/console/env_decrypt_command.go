package console

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"os"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
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
func (r *EnvDecryptCommand) Handle(ctx console.Context) (err error) {
	key := ctx.Option("key")
	if key == "" {
		ctx.Error("A decryption key is required.")
		return
	}
	encryptedData, err := os.ReadFile(".env.encrypted")
	if err != nil {
		ctx.Error("Encrypted environment file not found.")
		return
	}
	if _, err = os.Stat(".env"); err == nil {
		ok, _ := ctx.Confirm("Environment file already exists, are you sure to overwrite?", console.ConfirmOption{
			Default:     true,
			Affirmative: "Yes",
			Negative:    "No",
		})
		if !ok {
			return
		}
	}
	decrypted, err := decrypt(encryptedData, []byte(key))
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(".env", decrypted, 0644)
	if err != nil {
		panic(err)
	}
	ctx.Success("Encrypted environment successfully decrypted.")
	return nil
}

func decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, errors.New("A decryption key is required")
	}
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
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)
	return pkcs7Unpad(plaintext)
}

func pkcs7Unpad(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("empty data")
	}
	padding := int(data[length-1])
	if padding < 1 || padding > aes.BlockSize {
		return nil, errors.New("invalid padding")
	}
	if length < padding {
		return nil, errors.New("data shorter than padding")
	}
	return data[:length-padding], nil
}
