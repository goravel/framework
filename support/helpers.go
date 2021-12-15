package support

import (
	"log"
	"os"
)

func CreateEnv() {
	file, err := os.Create(".env")
	defer func() {
		err = file.Close()
		if err != nil {
			log.Fatalln("Close file fail:", err.Error())
		}
	}()

	if err != nil {
		log.Fatalln(err.Error())
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
	if err != nil {
		log.Fatalln(err.Error())
	}
}
