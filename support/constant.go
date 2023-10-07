package support

const Version string = "v1.13.2"

const (
	EnvRuntime = "runtime"
	EnvArtisan = "artisan"
	EnvTest    = "test"
)

var (
	Env          = EnvRuntime
	EnvPath      = ".env"
	RelativePath string
	RootPath     string
)
