package console

type Stubs struct{}

func (r Stubs) Channel() string {
	return `package DummyPackage

import (
	"github.com/goravel/framework/contracts/broadcasting"
)

func DummyChannel(user any, channelName string, params map[string]string) bool {
	return false
}

var _ broadcasting.ChannelAuthFunc = DummyChannel
`
}
