package support

var (
	RelativePath = ""
	RootPath     = ""

	RuntimeMode = ""

	EnvFilePath          = ".env"
	EnvFileEncryptPath   = ".env.encrypted"
	EnvFileEncryptCipher = "AES-256-CBC"

	DontVerifyEnvFileExists    = false
	DontVerifyEnvFileWhitelist = []string{"key:generate", "env:decrypt"}
)
