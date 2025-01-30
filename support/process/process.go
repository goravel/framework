package process

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/goravel/framework/support/str"
)

func Run(command string) (string, error) {
	cmd := exec.Command("/bin/sh", "-c", command)

	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s: %s", err, stderr.String())
	}

	return str.Of(out.String()).Squish().String(), nil
}
