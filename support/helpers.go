package support

import (
	"bufio"
	"github.com/goravel/framework/support/facades"
	"io"
	"os"
)

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

//GetLineNum Get file line num.
func GetLineNum(file string) int {
	total := 0
	f, _ := os.OpenFile(file, os.O_RDONLY, 0444)
	buf := bufio.NewReader(f)

	for {
		_, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
		} else {
			total++
		}
	}

	defer func() {
		f.Close()
	}()

	return total
}
