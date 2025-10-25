package console

type Stubs struct {
}

func (r Stubs) View() string {
	return `// DummyPathName
{{ define "DummyPathDefinition" }}
<h1>Welcome to DummyViewName</h1>
{{ end }}
`
}
