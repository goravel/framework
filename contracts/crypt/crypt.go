package crypt

//go:generate mockery --name=Crypt
type Crypt interface {
	// EncryptString encrypts the given string value, returning the encrypted string.
	EncryptString(value string) string
	// DecryptString decrypts the given string payload, returning the decrypted string.
	DecryptString(payload string) string
}
