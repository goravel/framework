package testing

import (
	"os"
	"strings"
)

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

REDIS_HOST=127.0.0.1
REDIS_PASSWORD=
REDIS_PORT=6379
`)
	return err
}
