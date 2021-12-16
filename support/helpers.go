package support

import (
	"bufio"
	"github.com/goravel/framework/support/facades"
	"io"
	"os"
	"strings"
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

func RunInTest() bool {
	return strings.HasSuffix(os.Args[0], ".test")
}

func CreateEnv() error {
	file, err := os.Create(".env")
	defer func() {
		file.Close()
	}()

	if err != nil {
		return err
	}

	_, err = file.WriteString(`APP_NAME=goravel
APP_ENV=local
APP_KEY=
APP_DEBUG=true
APP_URL=http://localhost
APP_HOST=127.0.0.1:3000

DB_CONNECTION=mysql
DB_HOST=
DB_PORT=3306
DB_DATABASE=
DB_USERNAME=
DB_PASSWORD=
`)
	return err
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