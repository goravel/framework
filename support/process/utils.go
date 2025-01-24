package process

import (
	"fmt"
	"net"

	"golang.org/x/exp/rand"
)

// Used by TestContainer, to simulate the port is using.
var TestPortUsing = false

func IsPortUsing(port int) bool {
	if TestPortUsing {
		return true
	}

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if l != nil {
		l.Close()
	}

	return err != nil
}

func ValidPort() int {
	for i := 0; i < 60; i++ {
		random := rand.Intn(10000) + 10000
		if !IsPortUsing(random) {
			return random
		}
	}

	return 0
}
