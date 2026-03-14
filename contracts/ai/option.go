package ai

const (
	OptionProvider = "provider"
	OptionModel    = "model"
	OptionTimeout  = "timeout"
)

// Option applies conversation options for provider selection and model behavior.
type Option func(map[string]any)
