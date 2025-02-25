package support

var (
	RelativePath = ""
	RootPath     = ""

	RuntimeMode = ""

	EnvFilePath            = ".env"
	EnvFileEncryptPath     = ".env.encrypted"
	EnvFileEncryptCipher   = "AES-256-CBC"
	EnvFileVerifyExists    = false
	EnvFileVerifyWhitelist = []string{"key:generate", "env:decrypt"}
)
