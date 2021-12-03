package facades

var Config ConfigFacade

type ConfigFacade interface {
	//Env Get config from env.
	Env(envName string, defaultValue ...interface{}) interface{}
	//Add Add config to application.
	Add(name string, configuration map[string]interface{})
	//Get Get config from application.
	Get(path string, defaultValue ...interface{}) interface{}
	//GetString Get string type config from application.
	GetString(path string, defaultValue ...interface{}) string
	//GetInt Get int type config from application.
	GetInt(path string, defaultValue ...interface{}) int
	//GetBool Get bool type config from application.
	GetBool(path string, defaultValue ...interface{}) bool
}