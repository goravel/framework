package process

import (
	"fmt"
	"math/rand"
	"net"

	"github.com/goravel/framework/errors"
)

// Used by TestContainer, to simulate the port is using.
var TestPortUsing = false

func IsPortUsing(port int) bool {
	if TestPortUsing {
		return true
	}

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if l != nil {
		errors.Ignore(l.Close)
	}

	return err != nil
}

func ValidPort() int {
	for range 60 {
		random := rand.Intn(10000) + 10000
		if !IsPortUsing(random) {
			return random
		}
	}

	return 0
}
