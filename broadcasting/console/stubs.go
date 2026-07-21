package console

type Stubs struct{}

func (r Stubs) Channel() string {
	return `package DummyPackage

type DummyChannel struct{}

func (r *DummyChannel) Name() string {
	return "channel-name"
}

func (r *DummyChannel) Join(user any, params map[string]string) bool {
	return false
}
`
}
