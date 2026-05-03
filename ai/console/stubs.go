package console

type Stubs struct {
}

func (r Stubs) Agent() string {
	return `package DummyPackage

import "github.com/goravel/framework/contracts/ai"

type DummyAgent struct {
}

func (r *DummyAgent) Instructions() string {
	return ""
}

func (r *DummyAgent) Messages() []ai.Message {
	return nil
}

func (r *DummyAgent) Middleware() []ai.Middleware {
	return nil
}

func (r *DummyAgent) Tools() []ai.Tool {
	return nil
}
`
}
