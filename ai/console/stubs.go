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

func (r Stubs) Tool() string {
	return `package DummyPackage

import "context"

type DummyTool struct {
}

func (r *DummyTool) Name() string {
	return "DummyName"
}

func (r *DummyTool) Description() string {
	return "A description of the tool."
}

func (r *DummyTool) Parameters() map[string]any {
	return nil
}

func (r *DummyTool) Execute(ctx context.Context, args map[string]any) (string, error) {
	return "", nil
}
`
}
