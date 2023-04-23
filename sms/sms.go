package sms

import (
	"fmt"
)

type Sms struct {
}

func (s *Sms) Send() bool {
	fmt.Println("Send")

	return false
}
