package ai

type Config struct {
	Default   string                    `json:"default"`
	Providers map[string]ProviderConfig `json:"providers"`
}

type ProviderConfig struct {
	Key      string                      `json:"key"`
	Models   ModelsConfig                `json:"models"`
	Url      string                      `json:"url"`
	Via      any                         `json:"via"` // Provider or func() (Provider, error)
	Failover map[FailoverReason][]string `json:"failover"`
}

type ModelsConfig struct {
	Text struct {
		Default   string `json:"default"`
		MaxTokens int    `json:"max_tokens"`
	} `json:"text"`
	Audio struct {
		Default string `json:"default"`
	} `json:"audio"`
	Transcription struct {
		Default string `json:"default"`
	} `json:"transcription"`
	Image struct {
		Default string `json:"default"`
	} `json:"image"`
}
