package ai

type Config struct {
	Default     string                    `json:"default"`
	Providers   map[string]ProviderConfig `json:"providers"`
	Attachments AttachmentConfig          `json:"attachments"`
}

type AttachmentConfig struct {
	MaxBytes int `json:"max_bytes"`
}

type ProviderConfig struct {
	Key    string       `json:"key"`
	Models ModelsConfig `json:"models"`
	Url    string       `json:"url"`
	Via    any          `json:"via"` // Provider or func() (Provider, error)
}

type ModelsConfig struct {
	Text struct {
		Default string `json:"default"`
	} `json:"text"`
}
