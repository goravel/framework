package ai

type Config struct {
	Default   string
	Providers map[string]ProviderConfig
}

type ProviderConfig struct {
	Key    string
	Models ModelsConfig
	Url    string
	Via    any // Provider or func() (Provider, error)
}

type ModelsConfig struct {
	Text struct {
		Default string
	}
}
