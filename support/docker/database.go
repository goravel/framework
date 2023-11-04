package docker

import (
	"fmt"
	"os"
	"os/exec"
)

type Database struct {
	file *os.File
}

func NewDatabase() *Database {
	return &Database{}
}

func (r *Database) Run() error {
	file, err := os.CreateTemp("", "goravel-docker-composer-*.yml")
	defer file.Close()

	r.file = file
	if err := os.WriteFile(file.Name(), []byte(Compose{}.Database()), 0755); err != nil {
		return err
	}

	cmd := fmt.Sprintf("docker-compose -f %s up --detach --quiet-pull", file.Name())
	_, err = exec.Command("/bin/sh", "-c", cmd).Output()

	return err
}

func (r *Database) Stop() error {
	cmd := fmt.Sprintf("docker-compose -f %s down", r.file.Name())
	_, err := exec.Command("/bin/sh", "-c", cmd).Output()
	defer os.Remove(r.file.Name())

	return err
}
