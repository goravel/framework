package support

import (
	"strings"
)

type Paths struct {
	// The base directory path, default is "app".
	App string
	// The bootstrap directory path, default is "bootstrap".
	Bootstrap string
	// The command directory path, default is "app/console/commands".
	Command string
	// The config directory path, default is "config".
	Config string
	// The controller directory path, default is "app/http/controllers".
	Controller string
	// The database directory path, default is "database".
	Database string
	// The event directory path, default is "app/events".
	Event string
	// The facades directory path, default is "app/facades".
	Facades string
	// The factory directory path, default is "database/factories".
	Factory string
	// The filter directory path, default is "app/filters".
	Filter string
	// The job directory path, default is "app/jobs".
	Job string
	// The language files directory path, default is "lang".
	Lang string
	// The listener directory path, default is "app/listeners".
	Listener string
	// The mail directory path, default is "app/mails".
	Mail string
	// The middleware directory path, default is "app/http/middleware".
	Middleware string
	// The migration directory path, default is "database/migrations".
	Migration string
	// The model directory path, default is "app/models".
	Model string
	// The observer directory path, default is "app/observers".
	Observer string
	// The package directory path, default is "packages".
	Package string
	// The policy directory path, default is "app/policies".
	Policy string
	// The provider directory path, default is "app/providers".
	Provider string
	// The public directory path, default is "public".
	Public string
	// The request directory path, default is "app/http/requests".
	Request string
	// The resources directory path, default is "resources".
	Resources string
	// The routes directory path, default is "routes".
	Routes string
	// The rule directory path, default is "app/rules".
	Rule string
	// The seeder directory path, default is "database/seeders".
	Seeder string
	// The storage directory path, default is "storage".
	Storage string
	// The test directory path, default is "tests".
	Test string
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
			App:        "app",
			Bootstrap:  "bootstrap",
			Command:    "app/console/commands",
			Config:     "config",
			Controller: "app/http/controllers",
			Database:   "database",
			Event:      "app/events",
			Facades:    "app/facades",
			Factory:    "database/factories",
			Filter:     "app/filters",
			Job:        "app/jobs",
			Lang:       "lang",
			Listener:   "app/listeners",
			Mail:       "app/mails",
			Middleware: "app/http/middleware",
			Migration:  "database/migrations",
			Model:      "app/models",
			Observer:   "app/observers",
			Package:    "packages",
			Policy:     "app/policies",
			Provider:   "app/providers",
			Public:     "public",
			Request:    "app/http/requests",
			Resources:  "resources",
			Routes:     "routes",
			Rule:       "app/rules",
			Seeder:     "database/seeders",
			Storage:    "storage",
			Test:       "tests",
		},
	}
)

func PathToSlice(path string) []string {
	path = strings.ReplaceAll(path, "\\", "/")
	path = strings.Trim(path, "/")
	if path == "" {
		return nil
	}

	return strings.Split(path, "/")
}

func PathPackage(pkg, def string) string {
	s := PathToSlice(pkg)

	if len(s) == 0 {
		return def
	}

	return s[len(s)-1]
}
