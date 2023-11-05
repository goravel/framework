package docker

import (
	"fmt"
	"os"
	"os/exec"
)

var usingDatabaseNum = 0
var usingDatabaseCompose *os.File

type DB struct {
}

func NewDB() *DB {
	return &DB{}
}

func (r *DB) Run() error {
	if usingDatabaseNum > 0 {
		usingDatabaseNum++

		return nil
	}

	file, err := os.CreateTemp("", "goravel-docker-composer-*.yml")
	if err != nil {
		return err
	}
	defer file.Close()
	usingDatabaseCompose = file

	if err := os.WriteFile(file.Name(), []byte(Compose{}.Database()), 0755); err != nil {
		return err
	}

	cmd := fmt.Sprintf("docker-compose -f %s up --detach --quiet-pull", file.Name())
	//cmd := fmt.Sprintf(`docker run -e "ACCEPT_EULA=Y" -e "MSSQL_SA_PASSWORD=yourStrong()" -p 1433:1433 --network dnmp-new_default --name mssql -d mysql:latest`)
	if o, err := exec.Command("/bin/sh", "-c", cmd).Output(); err != nil {
		fmt.Println(o, err.Error())
		return err
	}

	usingDatabaseNum++

	return nil
}

func (r *DB) Stop() error {
	usingDatabaseNum--
	if usingDatabaseNum > 0 {
		return nil
	}

	cmd := fmt.Sprintf("docker-compose -f %s down", usingDatabaseCompose.Name())
	_, err := exec.Command("/bin/sh", "-c", cmd).Output()
	defer func() {
		os.Remove(usingDatabaseCompose.Name())
		usingDatabaseCompose = nil
	}()

	return err
}
