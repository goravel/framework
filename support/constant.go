package support

const Version string = "v1.12.7"

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
