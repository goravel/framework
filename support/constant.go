package support

const Version string = "v1.12.5"

const (
	EnvRuntime = "runtime"
	EnvArtisan = "artisan"
	EnvTest    = "test"
)

var (
	Env      = EnvRuntime
	EnvPath  = ".env"
	RootPath string
)
