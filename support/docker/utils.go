package docker

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cast"

	"github.com/goravel/framework/contracts/testing"
)

func getExposedPort(exposedPorts []string, port int) int {
	for _, exposedPort := range exposedPorts {
		if !strings.Contains(exposedPort, cast.ToString(port)) {
			continue
		}

		ports := strings.Split(exposedPort, ":")

		return cast.ToInt(ports[0])
	}

	return 0
}

func getValidPort() int {
	for i := 0; i < 60; i++ {
		random := rand.Intn(10000) + 10000
		l, err := net.Listen("tcp", fmt.Sprintf(":%s", strconv.Itoa(random)))
		if err != nil {
			continue
		}
		defer l.Close()

		return random
	}

	return 0
}

func imageToCommand(image *testing.Image) (command string, exposedPorts []string) {
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
				port = fmt.Sprintf("%d:%s", getValidPort(), port)
			}
			ports = append(ports, port)
			commands = append(commands, "-p", port)
		}
	}

	commands = append(commands, fmt.Sprintf("%s:%s", image.Repository, image.Tag))

	return strings.Join(commands, " "), ports
}

func run(command string) (string, error) {
	cmd := exec.Command("/bin/sh", "-c", command)

	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s: %s", err, stderr.String())
	}

	return out.String(), nil
}
