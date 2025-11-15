package support

import "path/filepath"

type Paths struct {
	App        string
	Command    string
	Controller string
	Event      string
	Factory    string
	Filter     string
	Job        string
	Listener   string
	Mail       string
	Middleware string
	Migration  string
	Model      string
	Observer   string
	Package    string
	Policy     string
	Provider   string
	Request    string
	Rule       string
	Seeder     string
	Test       string
}

type Configuration struct {
	Paths Paths
}

var (
	RelativePath = ""
	RootPath     = ""

	RuntimeMode = ""

	EnvFilePath          = ".env"
	EnvFileEncryptPath   = ".env.encrypted"
	EnvFileEncryptCipher = "AES-256-CBC"

	DontVerifyEnvFileExists    = false
	DontVerifyEnvFileWhitelist = []string{"key:generate", "env:decrypt"}

	Config = Configuration{
		Paths: Paths{
			App:        filepath.Join("bootstrap", "app.go"),
			Command:    filepath.Join("app", "console", "commands"),
			Controller: filepath.Join("app", "http", "controllers"),
			Event:      filepath.Join("app", "events"),
			Factory:    filepath.Join("database", "factories"),
			Filter:     filepath.Join("app", "filters"),
			Job:        filepath.Join("app", "jobs"),
			Listener:   filepath.Join("app", "listeners"),
			Mail:       filepath.Join("app", "mails"),
			Middleware: filepath.Join("app", "http", "middleware"),
			Migration:  filepath.Join("database", "migrations"),
			Model:      filepath.Join("app", "models"),
			Observer:   filepath.Join("app", "observers"),
			Package:    "packages",
			Policy:     filepath.Join("app", "policies"),
			Provider:   filepath.Join("app", "providers"),
			Request:    filepath.Join("app", "http", "requests"),
			Rule:       filepath.Join("app", "rules"),
			Seeder:     filepath.Join("database", "seeders"),
			Test:       "tests",
		},
	}
)
