package support

const Version string = "v1.15.3"

const (
	EnvRuntime = "runtime"
	EnvArtisan = "artisan"
	EnvTest    = "test"
)

var (
	RelativePath = ""
	RootPath     = ""

	Env                    = EnvRuntime
	EnvFilePath            = ".env"
	EnvFileEncryptPath     = ".env.encrypted"
	EnvFileEncryptCipher   = "AES-256-CBC"
	EnvFileVerifyExists    = false
	EnvFileVerifyWhitelist = []string{"key:generate", "env:decrypt"}
)
