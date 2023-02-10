package crypt

//go:generate mockery --name=Crypt
type Crypt interface {
	// EncryptString encrypts the given string, returning the initialization vector and the encrypted string.
	EncryptString(value string) (string, string)
	// DecryptString decrypts the given string, returning the decrypted string.
	DecryptString(iv string, ciphertext string) string
}
