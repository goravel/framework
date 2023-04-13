package config

//go:generate mockery --name=Config
type Config interface {
	//Env Get config from env.
	Env(envName string, defaultValue ...any) any
	//Add config to application.
	Add(name string, configuration any)
	//Get config from application.
	Get(path string, defaultValue ...any) any
	//GetString Get string type config from application.
	GetString(path string, defaultValue ...any) string
	//GetInt Get int type config from application.
	GetInt(path string, defaultValue ...any) int
	//GetBool Get bool type config from application.
	GetBool(path string, defaultValue ...any) bool
}
