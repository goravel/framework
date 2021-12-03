package helpers

import "github.com/goravel/framework/support/facades"

//GetDatabaseConfig Get database config from ENV.
func GetDatabaseConfig() map[string]string {
	return map[string]string{
		"host":     facades.Config.GetString("database.connections." + facades.Config.GetString("database.default") + ".host"),
		"port":     facades.Config.GetString("database.connections." + facades.Config.GetString("database.default") + ".port"),
		"database": facades.Config.GetString("database.connections." + facades.Config.GetString("database.default") + ".database"),
		"username": facades.Config.GetString("database.connections." + facades.Config.GetString("database.default") + ".username"),
		"password": facades.Config.GetString("database.connections." + facades.Config.GetString("database.default") + ".password"),
		"charset":  facades.Config.GetString("database.connections." + facades.Config.GetString("database.default") + ".charset"),
	}
}
