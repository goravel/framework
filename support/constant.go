package support

const Version string = "v1.15.11"

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
