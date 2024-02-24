package support

const Version string = "v1.13.7"

const (
	EnvRuntime = "runtime"
	EnvArtisan = "artisan"
	EnvTest    = "test"
)

var (
	Env                  = EnvRuntime
	EnvPath              = ".env"
	IsKeyGenerateCommand = false
	RelativePath         string
	RootPath             string
)
