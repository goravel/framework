package support

const Version string = "v1.15.3"

const (
	EnvRuntime = "runtime"
	EnvArtisan = "artisan"
	EnvTest    = "test"
)

var (
	Env              = EnvRuntime
	EnvPath          = ".env"
	EnvEncryptPath   = ".env.encrypted"
	EnvEncryptCipher = "AES-256-CBC"
	RelativePath     = ""
	RootPath         = ""
)

var (
	EnvVerifyWhitelist []string
)
