package docker

import (
	"fmt"
	"strings"

	"github.com/goravel/framework/contracts/testing/docker"
	"github.com/goravel/framework/support/process"
)

func ExposedPort(exposedPorts []string, port string) string {
	for _, exposedPort := range exposedPorts {
		splitedExposedPort := strings.Split(exposedPort, ":")
		if len(splitedExposedPort) != 2 {
			continue
		}

		if splitedExposedPort[1] != port && !strings.Contains(splitedExposedPort[1], port+"/") {
			continue
		}

		return splitedExposedPort[0]
	}

	return ""
}

func ImageToCommand(image *docker.Image) (command string, exposedPorts []string) {
	if image == nil {
		return "", nil
	}

	commands := []string{"docker", "run", "--rm", "-d"}
	if len(image.Env) > 0 {
		for _, env := range image.Env {
			commands = append(commands, "-e", env)
		}
	}

	var ports []string
	if len(image.ExposedPorts) > 0 {
		for _, port := range image.ExposedPorts {
			if !strings.Contains(port, ":") {
				port = fmt.Sprintf("%d:%s", process.ValidPort(), port)
			}
			ports = append(ports, port)
			commands = append(commands, "-p", port)
		}
	}

	commands = append(commands, fmt.Sprintf("%s:%s", image.Repository, image.Tag))

	if len(image.Args) > 0 {
		commands = append(commands, image.Args...)
	}

	return strings.Join(commands, " "), ports
}
