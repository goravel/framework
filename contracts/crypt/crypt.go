package crypt

//go:generate mockery --name=Crypt
type Crypt interface {
	// EncryptString encrypts the given string value, returning the encrypted string and an error if any.
	EncryptString(value string) (string, error)
	// DecryptString decrypts the given string payload, returning the decrypted string and an error if any.
	DecryptString(payload string) (string, error)
}
